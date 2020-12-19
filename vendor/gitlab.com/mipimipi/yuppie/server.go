package yuppie

import (
	"context"
	"net/http"
	"runtime"
	"strings"
	"sync"

	"github.com/pkg/errors"
	l "github.com/sirupsen/logrus"
	utils "gitlab.com/mipimipi/go-utils"
	"gitlab.com/mipimipi/yuppie/desc"
	"gitlab.com/mipimipi/yuppie/internal/events"
	"gitlab.com/mipimipi/yuppie/internal/ssdp"
	"gitlab.com/mipimipi/yuppie/internal/types"
)

var log *l.Entry

func init() {
	log = l.WithFields(l.Fields{"srv": "upnp:srv"})
}

// URL paths
const (
	deviceDescPath      = "/device/devicedesc.xml" // device description
	deviceIconPath      = "/device/"               // device icons
	serviceDescPath     = "/services/desc/"        // service descriptions
	serviceControlPath  = "/services/control/"     // service control
	serviceEventSubPath = "/services/eventSub/"    // event subscriptions
)

// Server represents the UPnP server
type Server struct {
	cfg                 Config
	Errs                chan error
	Device              *rootDevice
	services            serviceMap
	bootID              *types.BootID
	configID            *types.ConfigID
	ssdps               []*ssdp.Server
	http                *http.Server
	presentationHandler func(http.ResponseWriter, *http.Request)
	httpHandlers        map[string](func(http.ResponseWriter, *http.Request))
	soapHandlers        map[string](func(map[string]StateVar) (SOAPRespArgs, SOAPError))
	evt                 *events.Eventing
	connected           bool
	// Locals contains variables that are persisted in the status.json of
	// yuppie
	Locals map[string]string
}

// New creates a new instance of the UPnP server from a device description and
// service descriptions.
// Note: The keys of the service map must correspond to the service ids in the
// device description
func New(cfg Config, rootDesc *desc.RootDevice, svcDescs desc.ServiceMap) (srv *Server, err error) {
	log.Trace("creating UPnP server ...")

	if cfg.equal(Config{}) {
		cfg = defaultCfg
	}

	// check that the device and service descriptions are OK
	if err = validateInputData(rootDesc, svcDescs); err != nil {
		err = errors.Wrap(err, "cannot create UPnP server")
		log.Fatal(err)
		return
	}

	srv = new(Server)

	// create optimized device and service objects from the descriptions. As a
	// side effect it is evaluated if multicast eventing is required
	if srv.Device, srv.services, err = createFromDesc(
		rootDesc,
		svcDescs,
		func() chan events.StateVar { return srv.evt.Listener },
	); err != nil {
		err = errors.Wrap(err, "cannot create UPnP server")
		log.Fatal(err)
		return
	}

	srv.Errs = make(chan error)
	srv.cfg = cfg
	srv.bootID = types.NewBootID()
	srv.configID = new(types.ConfigID)
	srv.httpHandlers = make(map[string](func(http.ResponseWriter, *http.Request)))
	srv.soapHandlers = make(map[string](func(map[string]StateVar) (SOAPRespArgs, SOAPError)))
	srv.Locals = make(map[string]string)
	if err = srv.setStatus(); err != nil {
		err = errors.Wrap(err, "cannot create UPnP server")
		log.Fatal(err)
		return
	}

	// srv.evt can only be create after srv.bootID is created. Otherwise a dump
	// will occur if state variables are multicasted
	srv.evt, err = events.NewEventing(cfg.Interfaces, srv.bootID)
	if err != nil {
		err = errors.Wrap(err, "cannot create UPnP server")
		log.Fatal(err)
		return
	}

	// create SSDP servers (one for each network interface)
	if err = srv.createSSDPServers(); err != nil {
		err = errors.Wrap(err, "cannot create UPnP server")
		log.Fatal(err)
		return
	}

	log.Trace("UPnP server created")

	return
}

// BootID returns the current value of BOOTID.UPNP.ORG
func (me *Server) BootID() uint32 {
	return me.bootID.Val()
}

// ConfigID returns the current value of CONFIGID.UPNP.ORG
func (me *Server) ConfigID() uint32 {
	return me.configID.Val()
}

// Connect starts the SSDP processes and multicast eventing
func (me *Server) Connect(ctx context.Context) (err error) {
	// nothing to do if already connected
	if me.connected {
		log.Info("tried to connect though server is already connected")
		return
	}

	log.Trace("connecting ...")

	// shutdown presentation HTTP server
	_ = me.http.Shutdown(ctx)

	// create general HTTP server
	me.createHTTPServer()

	// start general HTTP server
	go func() {
		if err := me.http.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP ListenAndServe: %v", err)
			me.Errs <- err
			return
		}
	}()
	log.Trace("general http server started")

	// start SSDP servers
	for _, ssdp := range me.ssdps {
		if err = ssdp.Connect(); err != nil {
			err = errors.Wrap(err, "cannot connect UPnP server")
			log.Fatal(err)
			return
		}
	}
	log.Trace("SSDP servers connected")

	// increase BootID as required by UPnP Device Architecture 2.0 spec
	me.bootID.Incr()

	me.evt.Run()

	me.connected = true

	log.Trace("connected")
	return
}

// Disconnect stops the SSDP processes and the multicast eventing
func (me *Server) Disconnect(ctx context.Context) {
	// nothing to do if not connected
	if !me.connected {
		log.Info("tried to disconnect though server is not connected")
		return
	}

	log.Trace("disconnecting ...")

	me.stop(ctx)

	// create presentation HTTP server
	me.createPresentationServer()

	// start presentation HTTP server
	go func() {
		if err := me.http.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP ListenAndServe: %v", err)
			me.Errs <- err
			return
		}
	}()
	log.Trace("presentation HTTP server started")

	me.connected = false

	log.Trace("disconnected")
}

// Errors returns a receive-only channel for errors from the UPnP server
func (me *Server) Errors() <-chan error {
	return me.Errs
}

// Run starts the server. It can be stopped via the context ctx
func (me *Server) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer func() {
		me.stop(ctx)
		close(me.Errs)
		me.evt.RemoveAllSubs()
		wg.Done()
	}()

	// start eventing listener for state variable changes
	me.evt.Listen(ctx)

	// initial multicast eventing of state variables
	me.sendEvents()

	// create presentation HTTP server
	me.createPresentationServer()

	// start presentation HTTP server
	go func() {
		if err := me.http.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP ListenAndServe: %v", err)
			me.Errs <- err
			return
		}
	}()
	log.Trace("presentation HTTP server started")

	log.Trace("running ...")

	// wait for cancellation
	<-ctx.Done()
	log.Trace("received cancel event")
	if err := me.writeStatus(); err != nil {
		me.Errs <- err
		return
	}
}

// ServerString assembles the server string in the format
// "<OS>/<OS version> UPnP/<UPnP version> <product name>/<product version>"
func (me *Server) ServerString() (s string) {
	// OS
	s = strings.Title(runtime.GOOS)

	// OS version
	si, err := utils.Uname()
	if err != nil {
		log.Error(err)
	} else {
		s += "/" + si.Release
	}

	// UPnP version
	s += " UPnP/2.0"

	// product name and version
	if len(me.cfg.ProductName) > 0 && len(me.cfg.ProductVersion) > 0 {
		s += " " + me.cfg.ProductName + "/" + me.cfg.ProductVersion
	} else {
		if len(me.cfg.ProductName) > 0 {
			s += " " + me.cfg.ProductName
		} else if len(me.cfg.ProductVersion) > 0 {
			s += " " + me.cfg.ProductVersion
		}
	}

	return
}

// StateVariable returns the state variable svName of service svcID
func (me *Server) StateVariable(svcID, svName string) (StateVar, bool) {
	if _, exists := me.services[svcID]; !exists {
		return nil, false
	}
	sv, exists := me.services[svcID].stateVars[svName]
	return StateVar(sv.StateVar), exists
}

// HTTPHandleFunc is a wrapper around http.ServeMux.HandleFunc. It allowes to
// register handler functions for HTTP requests for given patterns
func (me *Server) HTTPHandleFunc(pattern string, handleFunc func(http.ResponseWriter, *http.Request)) {
	log.Tracef("set handle func for pattern '%s'", pattern)

	me.httpHandlers[pattern] = handleFunc
}

// PresentationHandleFunc sets the handler function for HTTP calls to the
// presentation url of the root device
func (me *Server) PresentationHandleFunc(handleFunc func(http.ResponseWriter, *http.Request)) {
	log.Tracef("set handle func for presentatio URL '%s'", me.Device.Desc.Device.PresentationURL)

	me.presentationHandler = handleFunc
}

// SOAPHandleFunc allows to register functions to handle UPnP SOAP requests.
// Such handlers are defined per service ID / action combination
func (me *Server) SOAPHandleFunc(svcID string, act string, handler func(map[string]StateVar) (SOAPRespArgs, SOAPError)) {
	me.soapHandlers[svcID+"#"+act] = handler
}

// sendEvents traverses through the device tree and sends an initial event for
// each to-be-multicasted state variables
func (me *Server) sendEvents() {
	log.Trace("sending initial events ...")

	var sendEvents func(*device)
	sendEvents = func(dvc *device) {
		for _, svc := range dvc.services {
			for _, sv := range svc.stateVars {
				sv.SendEvent()
			}
		}
		for _, d := range dvc.devices {
			sendEvents(d)
		}
	}

	sendEvents(me.Device.device)

	log.Trace("initial events sent")
}

// serDescPaths sets the URL paths for service descriptions, service control
// and event subscription in the device description
func (me *Server) setDescPaths() {
	var setDescPaths func(*desc.Device)

	setDescPaths = func(dvc *desc.Device) {
		// process service info
		// note: this for-loop-variant had to be chosen since
		//       "for ... range ..." copies the array items which does not
		//       allow to change array items
		svcRefs := &(dvc.Services)
		for i := 0; i < len(*svcRefs); i++ {
			id := serviceID((*svcRefs)[i].ServiceID)

			// set service URLs. These are relative URLs.
			(*svcRefs)[i].SCPDURL = serviceDescPath + id.tail() + ".xml"
			(*svcRefs)[i].ControlURL = serviceControlPath + id.tail()
			(*svcRefs)[i].EventSubURL = serviceEventSubPath + id.tail()
		}

		// process embedded devices
		for i := 0; i < len(dvc.Devices); i++ {
			setDescPaths(&(dvc.Devices[i]))
		}
	}

	setDescPaths(&me.Device.Desc.Device)
}

// stop stop the SSDP processes, the eventing and the (general) HTTP server
func (me *Server) stop(ctx context.Context) {
	log.Trace("stopping ...")

	me.evt.Stop()

	// stop SSDP servers
	var wg sync.WaitGroup
	for _, ssdp := range me.ssdps {
		wg.Add(1)
		go ssdp.Disconnect(&wg)
	}
	wg.Wait()

	// shutdown general HTTP server
	_ = me.http.Shutdown(ctx)

	log.Trace("stopped")
}

// validateInputData checks if the device and service descriptions that were
// past to the server are OK. During that check, string values are trimmed.
func validateInputData(rootDesc *desc.RootDevice, svcDescs desc.ServiceMap) (err error) {
	// validate root device description
	ok, res := rootDesc.Validate()
	if !ok {
		err = errors.New(res[0])
		return
	}

	// validate service descriptions
	for _, svcDesc := range svcDescs {
		ok, res := svcDesc.Validate()
		if !ok {
			err = errors.New(res[0])
			return
		}
	}

	return
}
