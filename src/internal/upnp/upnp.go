package upnp

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	l "github.com/sirupsen/logrus"
	"gitlab.com/mipimipi/go-utils"
	"gitlab.com/mipimipi/muserv/src/internal/config"
	"gitlab.com/mipimipi/muserv/src/internal/content"
	"gitlab.com/mipimipi/yuppie"
	"gitlab.com/mipimipi/yuppie/desc"
)

// service IDs
const (
	svcIDContDir = "ContentDirectory"
	svcIDConnMgr = "ConnectionManager"
)

// names of state variables
const (
	svContainerUpdateIDs   = "ContainerUpdateIDs"
	svCurrentConnectionIDs = "CurrentConnectionIDs"
	svFeatureList          = "FeatureList"
	svServiceResetToken    = "ServiceResetToken"
	svSearchCapabilities   = "SearchCapabilities"
	svSortCapabilities     = "SortCapabilities"
	svSourceProtocolInfo   = "SourceProtocolInfo"
	svSystemUpdateID       = "SystemUpdateID"
)

// virtual command folder
const contentFolder = "/content/"

// content commands
const (
	albumsWithMultipleCovers = "albums-with-multiple-covers"
	inconsistentAlbums       = "inconsistent-albums"
	tracksWithoutAlbum       = "tracks-without-album"
	tracksWithoutCover       = "tracks-without-cover"
)

var log *l.Entry = l.WithFields(l.Fields{"srv": "upnp"})

// regular expression to check the right format of cover picture URLs
var rePictureURL = regexp.MustCompile(content.PictureFolder + `\d+\.jpg`)

// Server implements the muserv UPnP server
type Server struct {
	*yuppie.Server
	cfg config.Cfg
	cnt *content.Content
}

// New creates a new server instance
func New(ctx context.Context, cnt *content.Content) (upnp *Server, err error) {
	log.Trace("creating server ...")

	var srv *yuppie.Server

	// create yuppie UPnP server instance
	if srv, err = createUPnPServer(ctx); err != nil {
		return nil, errors.Wrap(err, "cannot create yuppie UPnP server")
	}

	upnp = &Server{
		srv,
		ctx.Value(config.KeyCfg).(config.Cfg),
		cnt,
	}

	upnp.InitStateVariables()

	// register handlers for request for presentation URL, music and picture
	// folder URLs
	upnp.setHTTPHandler()

	// register SOAP handlers
	upnp.setSOAPHandler()

	log.Trace("server created")

	return
}

// IncrSystemUpdateID increase state variable SystemUpdateID by count.
// exceeded is set to true if the maximum allowed value of SystemUpdateID
// was exceeded. In that case, the system reset procedure as described in the
// ContentDirectory service spec must be executed
func (me *Server) IncrSystemUpdateID(count uint32) (exceeded bool) {
	sv, exists := me.StateVariable(svcIDContDir, svSystemUpdateID)
	if !exists {
		err := fmt.Errorf("state variable '%s' not found: cannot increase", svSystemUpdateID)
		log.Fatal(err)
		me.Errs <- err
		return
	}
	sv.Lock()
	old := sv.Get().(uint32)
	if err := sv.Set(old + count); err != nil {
		err = errors.Wrapf(err, "cannot set state variable '%s' to %d", svSystemUpdateID, old+count)
		log.Fatal(err)
		me.Errs <- err
	}
	sv.Unlock()

	// if the new value is less than the old value, the range of system updated
	// id was exceeded
	exceeded = sv.Get().(uint32) < old

	log.Tracef("increased system update id to '%s'", sv.String())

	return
}

// InitStateVariables initializes all state variables
func (me *Server) InitStateVariables() {
	log.Trace("initializing state variables ...")

	// CurrentConnectionIDs
	sv, exists := me.StateVariable(svcIDConnMgr, svCurrentConnectionIDs)
	if !exists {
		err := fmt.Errorf("state variable '%s' not found: cannot initialize", svCurrentConnectionIDs)
		log.Fatal(err)
		me.Errs <- err
		return
	}
	// - since muserv does not implement the action PrepareForConnection(), the
	//   response is always "0" as required by ConnectionManager:2, Service
	//   Template Version 1.01
	sv.Lock()
	if err := sv.Init("0"); err != nil {
		err := errors.Wrapf(err, "cannot initialize state variable '%s'", svCurrentConnectionIDs)
		log.Fatal(err)
		me.Errs <- err
	}
	sv.Unlock()

	// SourceProtocolInfo
	sv, exists = me.StateVariable(svcIDConnMgr, svSourceProtocolInfo)
	if !exists {
		err := fmt.Errorf("state variable '%s' not found: cannot initialize", svSourceProtocolInfo)
		log.Fatal(err)
		me.Errs <- err
		return
	}
	// - set supported mime types
	sv.Lock()
	if sv.String() == "" {
		if err := sv.Init(config.SupportedMimeTypes()); err != nil {
			err = errors.Wrapf(err, "cannot initialize state variable '%s'", svSourceProtocolInfo)
			log.Fatal(err)
			me.Errs <- err
		}
	}
	sv.Unlock()

	// ServiceResetToken: make clients reset their buffers by giving service
	// reset token a new value
	me.SetServiceResetToken()

	// ContainerUpdateIDs
	me.SetContainerUpdateIDs("")

	// SystemUpdateID: initialize it with 0 if it's not set already
	sv, exists = me.StateVariable(svcIDContDir, svSystemUpdateID)
	if !exists {
		err := fmt.Errorf("state variable '%s' not found: cannot initialize", svSystemUpdateID)
		log.Fatal(err)
		me.Errs <- err
		return
	}
	sv.Lock()
	if sv.String() == "" {
		if err := sv.Init(uint32(0)); err != nil {
			err = errors.Wrapf(err, "cannot initialize state variable '%s'", svSystemUpdateID)
			log.Fatal(err)
			me.Errs <- err
		}
	}
	sv.Unlock()

	log.Trace("state variables initialized")
}

// ServiceResetProcedure executes the service reset procedure as described in
// the ContentDirectory service specification
func (me *Server) ServiceResetProcedure(ctx context.Context) {
	log.Trace("executing service reset procudure")
	me.Disconnect(ctx)
	me.SetServiceResetToken()
	me.SetContainerUpdateIDs("")
	me.cnt.ResetCtrUpdCounts()
	if err := me.Connect(ctx); err != nil {
		err = errors.Wrap(err, "cannot connect after service reset procedure")
		me.Errs <- err
	}
}

// SetContainerUpdateIDs set state variable ContainerUpdateIDs to updates
func (me *Server) SetContainerUpdateIDs(updates string) {
	sv, exists := me.StateVariable(svcIDContDir, svContainerUpdateIDs)
	if !exists {
		err := fmt.Errorf("state variable '%s' not found: cannot set", svServiceResetToken)
		log.Fatal(err)
		me.Errs <- err
		return
	}
	sv.Lock()
	if err := sv.Set(updates); err != nil {
		err = errors.Wrapf(err, "cannot set state variable '%s'", svContainerUpdateIDs)
		log.Fatal(err)
		me.Errs <- err
	}
	sv.Unlock()
	log.Tracef("set %s to %s", svContainerUpdateIDs, sv.String())
}

// SetServiceResetToken assigns a new random string to state variable
// ServiceResetToken
func (me *Server) SetServiceResetToken() {
	sv, exists := me.StateVariable(svcIDContDir, svServiceResetToken)
	if !exists {
		err := fmt.Errorf("state variable '%s' not found: cannot set", svServiceResetToken)
		log.Fatal(err)
		me.Errs <- err
		return
	}
	sv.Lock()
	if err := sv.Set(utils.RandomString(32)); err != nil {
		err := errors.Wrapf(err, "cannot set state variable '%s'", svServiceResetToken)
		log.Fatal(err)
		me.Errs <- err
	}
	sv.Unlock()
	log.Tracef("set state variable '%s' to '%s'", svServiceResetToken, sv.String())
}

// createUPnPServer create a new instance of the yuppie UPnP server
func createUPnPServer(ctx context.Context) (srv *yuppie.Server, err error) {
	log.Trace("creating yuppie UPnP server ...")

	// create configuration
	cfg := ctx.Value(config.KeyCfg).(config.Cfg)
	srvCfg := yuppie.Config{
		Interfaces:     cfg.UPnP.Interfaces,
		Port:           cfg.UPnP.Port,
		MaxAge:         cfg.UPnP.MaxAge,
		ProductName:    "muserv",
		ProductVersion: ctx.Value(config.KeyVersion).(string),
		StatusFile:     cfg.UPnP.StatusFile,
		IconRootDir:    config.IconDir,
	}

	// create root device
	root := desc.RootDevice{
		XMLName: xml.Name{
			Local: "root",
			Space: "urn:schemas-upnp-org:device-1-0",
		},
		SpecVersion: desc.SpecVersion{
			Major: 2,
			Minor: 0,
		},
		Device: desc.Device{
			DeviceType:       "urn:schemas-upnp-org:device:MediaServer:1",
			FriendlyName:     cfg.UPnP.ServerName,
			Manufacturer:     cfg.UPnP.Device.Manufacturer,
			ManufacturerURL:  cfg.UPnP.Device.ManufacturerURL,
			ModelDescription: cfg.UPnP.Device.ModelDescription,
			ModelName:        cfg.UPnP.Device.ModelName,
			ModelNumber:      cfg.UPnP.Device.ModelNumber,
			ModelURL:         cfg.UPnP.Device.ModelURL,
			SerialNumber:     cfg.UPnP.Device.SerialNumber,
			UDN:              "uuid:" + cfg.UPnP.UUID,
			UPC:              cfg.UPnP.Device.UPC,
			Icons: []desc.Icon{
				{
					Mimetype: "image/png",
					Width:    300,
					Height:   300,
					Depth:    8,
					URL:      "/icon_dark.png",
				},
				{
					Mimetype: "image/png",
					Width:    300,
					Height:   300,
					Depth:    8,
					URL:      "/icon_light.png",
				},
			},
			Services: []desc.ServiceReference{
				{
					ServiceType: "urn:schemas-upnp-org:service:ContentDirectory:4",
					ServiceID:   "urn:upnp-org:serviceId:" + svcIDContDir,
				},
				{
					ServiceType: "urn:schemas-upnp-org:service:ConnectionManager:2",
					ServiceID:   "urn:upnp-org:serviceId:" + svcIDConnMgr,
				},
			},
			PresentationURL: "/",
		},
	}

	// create service descriptions
	var svc *desc.Service
	svcs := make(desc.ServiceMap)
	// - ContentDirectory service
	svc, err = desc.LoadService(filepath.Join(config.CfgDir, svcIDContDir+".xml"))
	if err != nil {
		err = errors.Wrap(err, "cannot read description of ContentDirectory service")
		return
	}
	svcs[svcIDContDir] = svc
	// - ConnectionManager service
	svc, err = desc.LoadService(filepath.Join(config.CfgDir, svcIDConnMgr+".xml"))
	if err != nil {
		err = errors.Wrap(err, "cannot read description of ConnectionManager service")
		return
	}
	svcs[svcIDConnMgr] = svc

	if srv, err = yuppie.New(srvCfg, &root, svcs); err != nil {
		err = errors.Wrap(err, "cannot create yuppie UPnP server")
		return
	}

	log.Trace("yuppie UPnP server created")

	return
}

// setHTTPHandler set the handler for HTTP request for presentation URL, music
// and picture folder URLs
func (me *Server) setHTTPHandler() {
	stateVar := func(svName string) string {
		sv, exists := me.StateVariable(svcIDContDir, svName)
		if !exists {
			err := fmt.Errorf("state variable %s not found: cannot display", svName)
			log.Fatal(err)
			return ""
		}
		return fmt.Sprintf("    %s: %s\n", svName, sv.String())
	}

	// handler for presentation URL
	me.PresentationHandleFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "%s [%s]\n\n", me.cfg.UPnP.ServerName, me.Device.UDN[5:])
			fmt.Fprintf(w, "%s\n\n", me.ServerString())

			fmt.Fprint(w, "Status:\n")
			fmt.Fprintf(w, "    BOOTID.UPNP.ORG: %d\n", me.BootID())
			fmt.Fprintf(w, "    CONFIGID.UPNP.ORG: %d\n", me.ConfigID())
			fmt.Fprint(w, stateVar(svServiceResetToken))
			fmt.Fprint(w, stateVar(svSystemUpdateID))
			fmt.Fprintf(w, "%s\n", stateVar(svContainerUpdateIDs))

			me.cnt.WriteStatus(w)
		},
	)

	// handler for requests to music folder
	me.HTTPHandleFunc(content.MusicFolder,
		func(w http.ResponseWriter, r *http.Request) {
			log.Tracef("received request for music: %s", r.URL.String())

			path, err := url.QueryUnescape(r.URL.String())
			if err != nil {
				log.Errorf("cannot unescape URL: %s", r.URL.String())
				return
			}

			if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
				// if path is an external URI, the file is under that path, ...
				http.ServeFile(w, r, path)
			} else {
				// ... otherwise: serve the corresponding file from the music
				// directory
				path = path[len(content.MusicFolder):]
				dir := me.cfg.Cnt.MusicDir(path)
				if len(dir) == 0 {
					log.Errorf("requested file '%s' not found in any of the music directories", path)
					return
				}
				http.ServeFile(w, r, filepath.Join(dir, path))
			}
		},
	)

	// handler for requests to pictures folder
	me.HTTPHandleFunc(content.PictureFolder,
		func(w http.ResponseWriter, r *http.Request) {
			log.Tracef("received request for picture: %s", r.URL.String())

			path, err := url.QueryUnescape(r.URL.String())
			if err != nil {
				err = errors.Wrapf(err, "cannot unescape URL: %s", r.URL.String())
				log.Fatal(err)
				http.Error(w, fmt.Sprintf("server error: cannot unescape URL: %s", r.URL.String()), http.StatusInternalServerError)
			}
			// verify that path has required format (the picture file name is
			// "<PICTURE-ID>.jpg", where PICTURE-ID is int64)
			if !rePictureURL.MatchString(path) {
				log.Fatalf("mal-formed picture URL: %s", r.URL.String())
				http.Error(w, fmt.Sprintf("server error: mal-formed picture URL: %s", r.URL.String()), http.StatusInternalServerError)
				return
			}
			// retrieve int64 ID of requested picture
			id, err := strconv.ParseUint(path[len(content.PictureFolder):len(path)-4], 10, 64)
			if err != nil {
				log.Fatalf("cannot retrieve picture id from URL: %s", r.URL.String())
				http.Error(w, fmt.Sprintf("server error: cannot retrieve picture id from URL: %s", r.URL.String()), http.StatusInternalServerError)
				return
			}
			// get picture from picture map
			picture := me.cnt.Picture(id)
			if picture == nil {
				log.Errorf("picture with id %d is unknown", id)
				http.Error(w, fmt.Sprintf("server error: picture %d is unknown", id), http.StatusInternalServerError)
				return
			}
			// return picture
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Content-Length", strconv.Itoa(len(*picture)))
			if _, err := w.Write(*picture); err != nil {
				err = errors.Wrapf(err, "cannot write picture id %d to HTTP response", id)
				log.Fatal(err)
				http.Error(w, fmt.Sprintf("server error: cannot write picture id %d to HTTP response", id), http.StatusInternalServerError)
				return
			}
		},
	)

	// handler for command requests
	me.HTTPHandleFunc(contentFolder,
		func(w http.ResponseWriter, r *http.Request) {
			path, err := url.QueryUnescape(r.URL.String())
			if err != nil {
				err = errors.Wrapf(err, "cannot unescape URL: %s", r.URL.String())
				log.Fatal(err)
				http.Error(w, fmt.Sprintf("server error: cannot unescape URL: %s", r.URL.String()), http.StatusInternalServerError)
			}

			switch path[len(contentFolder):] {
			case albumsWithMultipleCovers:
				me.cnt.AlbumsWithMultipleCovers(w)
			case inconsistentAlbums:
				me.cnt.InconsistentAlbums(w)
			case tracksWithoutAlbum:
				me.cnt.TracksWithoutAlbum(w)
			case tracksWithoutCover:
				me.cnt.TracksWithoutCover(w)
			default:
				fmt.Fprint(w, "unknown command")
			}
		},
	)
}

// setSOAPHandler sets handler functions for SOAP actions of the
// ContentDirectory and the ConnectionManager services
func (me *Server) setSOAPHandler() {
	me.SOAPHandleFunc(svcIDContDir, "GetSearchCapabilities",
		func(reqArgs map[string]yuppie.StateVar) (yuppie.SOAPRespArgs, yuppie.SOAPError) {
			return me.getSearchCapabilities(reqArgs)
		})
	me.SOAPHandleFunc(svcIDContDir, "GetSortCapabilities",
		func(reqArgs map[string]yuppie.StateVar) (yuppie.SOAPRespArgs, yuppie.SOAPError) {
			return me.getSortCapabilities(reqArgs)
		})
	me.SOAPHandleFunc(svcIDContDir, "GetFeatureList",
		func(reqArgs map[string]yuppie.StateVar) (yuppie.SOAPRespArgs, yuppie.SOAPError) {
			return me.getFeatureList(reqArgs)
		})
	me.SOAPHandleFunc(svcIDContDir, "GetSystemUpdateID",
		func(reqArgs map[string]yuppie.StateVar) (yuppie.SOAPRespArgs, yuppie.SOAPError) {
			return me.getSystemUpdateID(reqArgs)
		})
	me.SOAPHandleFunc(svcIDContDir, "GetServiceResetToken",
		func(reqArgs map[string]yuppie.StateVar) (yuppie.SOAPRespArgs, yuppie.SOAPError) {
			return me.getServiceResetToken(reqArgs)
		})
	me.SOAPHandleFunc(svcIDContDir, "Browse",
		func(reqArgs map[string]yuppie.StateVar) (yuppie.SOAPRespArgs, yuppie.SOAPError) {
			return me.browse(reqArgs)
		})
	me.SOAPHandleFunc(svcIDConnMgr, "GetProtocolInfo",
		func(reqArgs map[string]yuppie.StateVar) (yuppie.SOAPRespArgs, yuppie.SOAPError) {
			return me.getProtocolInfo(reqArgs)
		})
	me.SOAPHandleFunc(svcIDConnMgr, "GetCurrentConnectionIDs",
		func(reqArgs map[string]yuppie.StateVar) (yuppie.SOAPRespArgs, yuppie.SOAPError) {
			return me.getCurrentConnectionIDs(reqArgs)
		})
	me.SOAPHandleFunc(svcIDConnMgr, "GetCurrentConnectionInfo",
		func(reqArgs map[string]yuppie.StateVar) (yuppie.SOAPRespArgs, yuppie.SOAPError) {
			return me.getCurrentConnectionInfo(reqArgs)
		})
}
