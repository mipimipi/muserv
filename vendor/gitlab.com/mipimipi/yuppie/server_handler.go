package yuppie

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	utils "gitlab.com/mipimipi/go-utils"
	"gitlab.com/mipimipi/go-utils/file"
	"gitlab.com/mipimipi/yuppie/internal/events"
)

const httpProtocol = "http"

// createPresentationServer creates a new HTTP server me.http. The server only serves
// the presentation URL of the root device
func (me *Server) createPresentationServer() {
	log.Trace("creating presentation server")

	// create HTTP server
	me.http = new(http.Server)
	if me.cfg.Port != 0 {
		me.http.Addr = ":" + strconv.Itoa(me.cfg.Port)
	}
	mux := http.NewServeMux()
	me.http.Handler = mux

	mux.HandleFunc(me.Device.Desc.Device.PresentationURL, me.presentationHandler)

	log.Tracef("presentation server created on %s", me.http.Addr)
}

// createHTTPServer creates the new HTTP server me.http. It serves the
// presentation URL of the root device and all other required URLs
func (me *Server) createHTTPServer() {
	log.Trace("creating HTTP server")

	me.createPresentationServer()
	me.setHTTPHandleFuncs()

	log.Tracef("HTTP server created on %s", me.http.Addr)
}

// setHttpHandleFuncs registers handler functions for device description,
// service description requests and other URL patterns
func (me *Server) setHTTPHandleFuncs() {
	// device description
	me.http.Handler.(*http.ServeMux).HandleFunc(deviceDescPath,
		func(w http.ResponseWriter, r *http.Request) {
			me.deviceDescHandler(w, r)
		},
	)

	// device icons
	me.http.Handler.(*http.ServeMux).HandleFunc(deviceIconPath,
		func(w http.ResponseWriter, r *http.Request) {
			me.deviceIconHandler(w, r)
		},
	)

	// service descriptions
	me.http.Handler.(*http.ServeMux).HandleFunc(serviceDescPath,
		func(w http.ResponseWriter, r *http.Request) {
			me.serviceDescHandler(w, r)
		},
	)

	// service control
	me.http.Handler.(*http.ServeMux).HandleFunc(serviceControlPath,
		func(w http.ResponseWriter, r *http.Request) {
			me.serviceControlHandler(w, r)
		},
	)

	// event subscription
	me.http.Handler.(*http.ServeMux).HandleFunc(serviceEventSubPath,
		func(w http.ResponseWriter, r *http.Request) {
			me.serviceEventSubHandler(w, r)
		},
	)

	// other patterns
	for pattern, handleFunc := range me.httpHandlers {
		me.http.Handler.(*http.ServeMux).HandleFunc(pattern, handleFunc)
	}
}

// deviceDescHandler handles requests for the device description, i.e. requests
// for /device/devicedesc.xml
func (me *Server) deviceDescHandler(w http.ResponseWriter, r *http.Request) {
	log.Trace("description.xml requested")

	// device description must contain current ConfigID as required by UPnP
	// Device Architecture 2.0
	me.Device.Desc.ConfigID = me.configID.Val()

	// set the paths for the service descriptions in the device description structures
	me.setDescPaths()

	// render service description
	desc, err := utils.MarshalXML(me.Device.Desc)
	if err != nil {
		log.Errorf("couldn't marshal device description: %v", err)
		http.Error(w, "can't create device description", http.StatusInternalServerError)
		return
	}
	// set header fields
	setHeader(w, me.ServerString(), len(desc))
	// as per UPnP Device Architecture 2.0 spec, the content language must be
	// contained in the response if and only if the request contained the
	// ACCEPT-LANGUAGE field
	if r.Header.Get("ACCEPT-LANGUAGE") != "" {
		w.Header().Set("content-language", "en-US")
	}

	// send response
	if _, err = w.Write(desc); err != nil {
		err = errors.Wrap(err, "couldn't send device description response")
		log.Fatal(err)
	}
}

// deviceIconHandler handles requests for device icons, i.e. requests
// for /device/*
func (me *Server) deviceIconHandler(w http.ResponseWriter, r *http.Request) {
	log.Tracef("icon requested: %s", r.URL.String())

	iconPath, err := url.QueryUnescape(r.URL.String())
	if err != nil {
		err = errors.Wrapf(err, "cannot unescape URL: %s", r.URL.String())
		log.Fatal(err)
		http.Error(w, fmt.Sprintf("server error: cannot unescape URL: %s", r.URL.String()), http.StatusInternalServerError)
	}

	// return icon
	http.ServeFile(w, r, filepath.Join(me.cfg.IconRootDir, iconPath[len(deviceIconPath):]))
}

// serviceDescHandler handles requests for service descriptions, i.e. requests
// for /upnp/services/desc/<service-id>.xml. service-id is the actual id of a
// service, i.e. it's the <service-id> part of the service id field of the
// device description: urn:<domain>:serviceid:<service-id>
func (me *Server) serviceDescHandler(w http.ResponseWriter, r *http.Request) {
	// extract id of requested service. That's the filename of requested
	// description without ".xml" suffix
	_, filename := path.Split(r.URL.Path)
	id := file.PathTrunk(filename)

	log.Tracef("service description for %s requested", id)

	// get service
	svc, ok := me.services[id]
	if !ok {
		log.Fatalf("service with id '%s' couldn't be found", id)
		http.Error(w, fmt.Sprintf("service '%s' is unknown", id), http.StatusInternalServerError)
		return
	}

	// service description must contain current ConfigID as required by UPnP
	// Device Architecture 2.0
	svc.desc.ConfigID = me.configID.Val()

	// render service description
	desc, err := utils.MarshalXML(svc.desc)
	if err != nil {
		err = errors.Wrap(err, "cannot render service description")
		log.Fatal(err)
		http.Error(w, "can't create service description", http.StatusInternalServerError)
		return
	}

	// contained in the response if and only if the request contained the
	// ACCEPT-LANGUAGE field
	if r.Header.Get("ACCEPT-LANGUAGE") != "" {
		w.Header().Set("content-language", "en-US")
	}

	// send response
	if _, err = w.Write(desc); err != nil {
		err = errors.Wrap(err, "couldn't send device description response")
		log.Fatal(err)
	}
}

// serviceControlHandler handles requests that represent calls of SOAP actions
func (me *Server) serviceControlHandler(w http.ResponseWriter, r *http.Request) {
	log.Trace("service control request received")

	// get SOAP action
	svcID, actName, err := me.parseSOAPAction(w, r)
	if err != nil {
		err = errors.Wrap(err, "cannot parse SOAP action")
		log.Error(err)
		return
	}

	// verify that a handler for that action exists
	handler, exists := me.soapHandlers[svcID+"#"+actName]
	if !exists {
		me.sendSOAPFault(w,
			SOAPError{
				Code: UPnPErrorOptActionNotImplemented,
				Desc: fmt.Sprintf("no handler for action '%s'", svcID+"#"+actName),
			},
		)
		log.Errorf("no handler for action '%s'", svcID+"#"+actName)
		return
	}

	// get input arguments of action and verify that all of them fulfill the
	// conditions wrt. ranges and allowed values
	args, err := me.parseSOAPArguments(w, r, svcID, actName)
	if err != nil {
		err = errors.Wrap(err, "cannot parse SOAP arguments")
		log.Error(err)
		return
	}

	// invoke handler function and receive the response arguments
	argsOut, soapErr := handler(args)
	if !soapErr.IsNil() {
		me.sendSOAPFault(w, soapErr)
		return
	}

	// render response SOAP document
	// note: Existence of service has already been check in parseSOAPAction()
	svc := me.services[svcID]
	respBody := soapActResp{
		serviceType: string(svc.typ),
		serviceVer:  string(svc.ver),
		name:        actName,
		args:        argsOut,
	}.marshal()

	// send response
	setHeader(w, me.ServerString(), len(respBody))
	_, err = w.Write(respBody)
	if err != nil {
		log.Error(err)
	}
}

// serviceEventSubHandler handles event subscription requests
func (me *Server) serviceEventSubHandler(w http.ResponseWriter, r *http.Request) {
	log.Tracef("event %s request received: ", r.Method)

	switch r.Method {
	case "SUBSCRIBE":
		if r.Header.Get("SID") == "" {
			// new subscription

			// check value of NT
			if r.Header.Get("NT") != "upnp:event" {
				log.Error("server error: NT is not 'upnp:event'")
				http.Error(w, "NT is not 'upnp:event'", http.StatusPreconditionFailed)
				return
			}

			// check retrieve delivery urls
			urls, err := events.ParseURLs(r.Header.Get("CALLBACK"))
			if err != nil {
				err = errors.Wrapf(err, "cannot parse callback url(s) from '%s'", r.Header.Get("CALLBACK"))
				log.Error(err)
				http.Error(w, "invalid callback url(s)", http.StatusPreconditionFailed)
				return
			}

			// retrieve desired subscription duration
			dur, err := events.ParseTimeout(r.Header.Get("TIMEOUT"))
			if err != nil {
				err = errors.Wrapf(err, "could not parse timeout from '%s'", r.Header.Get("TIMEOUT"))
				log.Error(err)
				http.Error(w, "invalid TIMEOUT", http.StatusPreconditionFailed)
				return
			}

			// determine state variables that are evented
			var svs []events.StateVar
			for _, svc := range me.services {
				for _, sv := range svc.stateVars {
					if sv.toBeEvented {
						svs = append(svs, sv)
					}
				}
			}

			// assemble and send response
			sid := me.evt.AddSub(dur, urls, svs)
			w.Header().Set("DATE", time.Now().Format(time.RFC1123))
			w.Header().Set("SERVER", me.ServerString())
			w.Header().Set("SID", "uuid:"+sid.String())
			w.Header().Set("CONTENT-LENGTH", "0")
			w.Header().Set("TIMEOUT", "Second-"+fmt.Sprintf("%d", int(dur.Seconds())))
			w.WriteHeader(http.StatusOK)
		} else {
			// renewal of existing subscription

			// neither NT nor CALLBACK must be set
			if r.Header.Get("NT") != "" || r.Header.Get("CALLBACK") != "" {
				log.Error("precondition failed: neither NT nor CALLBACK must be set")
				http.Error(w, "precondition failed: neither NT nor CALLBACK must be set", http.StatusBadRequest)
				return
			}

			// retrieve desired subscription duration
			dur, err := events.ParseTimeout(r.Header.Get("TIMEOUT"))
			if err != nil {
				err = errors.Wrapf(err, "could not parse timeout from '%s'", r.Header.Get("TIMEOUT"))
				log.Error(err)
				http.Error(w, "could not parse TIMEOUT", http.StatusPreconditionFailed)
				return
			}

			if err := me.evt.RenewSub(uuid.MustParse(r.Header.Get("SID")[5:]), dur); err != nil {
				err = errors.Wrapf(err, "SID %s not found - unable to accept renewal", r.Header.Get("SID"))
				log.Error(err)
				http.Error(w, fmt.Sprintf("SID %s not found - unable to accept renewal", r.Header.Get("SID")), http.StatusInternalServerError)
				return
			}

			// assemble and send response
			w.Header().Set("DATE", time.Now().Format(time.RFC1123))
			w.Header().Set("SERVER", me.ServerString())
			w.Header().Set("SID", r.Header.Get("SID"))
			w.Header().Set("CONTENT-LENGTH", "0")
			w.Header().Set("TIMEOUT", "Second-"+fmt.Sprintf("%d", int(dur.Seconds())))
			w.WriteHeader(http.StatusOK)
		}
	case "UNSUBSCRIBE":
		// unsubscribe
		if err := me.evt.RemoveSub(uuid.MustParse(r.Header.Get("SID")[5:])); err != nil {
			err = errors.Wrapf(err, "SID %s not found - unable to unsubscribe", r.Header.Get("SID"))
			log.Error(err)
			http.Error(w, fmt.Sprintf("SID %s not found - unable to unsubscribe", r.Header.Get("SID")), http.StatusInternalServerError)
			return
		}

		// send response
		w.WriteHeader(http.StatusOK)

	default:
		log.Errorf("server error: unknown method '%s'", r.Method)
		http.Error(w, fmt.Sprintf("unknown method '%s'", r.Method), http.StatusMethodNotAllowed)
		return
	}
}

func setHeader(w http.ResponseWriter, server string, n int) {
	w.Header().Set("server", server)
	w.Header().Set("date", time.Now().Format(time.RFC1123))
	w.Header().Set("content-type", "text/xml; charset=\"utf-8\"")
	w.Header().Set("content-length", fmt.Sprint(n))
}
