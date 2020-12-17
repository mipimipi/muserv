package yuppie

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var reSOAPAction = regexp.MustCompile(`"urn:schemas-upnp-org:service:.+:.+#.+"`)

// parseSOAPAction evaluates a HTTP request to check if it's a call of a SOAP
// action of a. If that's the case, the corresponding service id and action
// name is extracted
func (me *Server) parseSOAPAction(w http.ResponseWriter, r *http.Request) (id, act string, err error) {
	// extract service id and check if that it's valid
	_, id = path.Split(r.URL.Path)
	if _, exists := me.services[id]; !exists {
		err = fmt.Errorf("service '%s' does not exist", id)
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check if requested SOAPACTION is valid
	soapAct := r.Header.Get("SOAPACTION")
	if !reSOAPAction.MatchString(soapAct) {
		me.sendSOAPFault(w,
			SOAPError{
				Code: UPnPErrorInvalidAction,
				Desc: fmt.Sprintf("invalid SOAPACTION: %s", soapAct),
			},
		)
		err = fmt.Errorf("invalid SOAPACTION: %s", soapAct)
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s := strings.Split(soapAct[1:len(soapAct)-1], "#")
	// - does requested action exist?
	if _, exists := me.services[id].actSpecs[s[1]]; !exists {
		me.sendSOAPFault(w,
			SOAPError{
				Code: UPnPErrorInvalidAction,
				Desc: fmt.Sprintf("unknown action: %s", s[1]),
			},
		)
		err = fmt.Errorf("unknown action: '%s'", s[1])
		log.Error(err)
		return
	}
	act = s[1]
	// - service type and version OK?
	s = strings.Split(s[0], ":")
	svcTyp := strings.Join(s[0:4], ":")
	svcVer := s[4]
	if me.services[id].typ != serviceType(svcTyp) {
		me.sendSOAPFault(w,
			SOAPError{
				Code: UPnPErrorInvalidAction,
				Desc: fmt.Sprintf("unknown service type: %s", svcTyp),
			},
		)
		err = fmt.Errorf("unknown service type: '%s'", svcTyp)
		log.Error(err)
		return
	}
	if me.services[id].ver < serviceVersion(svcVer) {
		me.sendSOAPFault(w,
			SOAPError{
				Code: UPnPErrorInvalidAction,
				Desc: fmt.Sprintf("requested service version too high: %s", svcVer),
			},
		)
		err = fmt.Errorf("requested service version too high: %s", svcVer)
		log.Error(err)
		return
	}

	return
}

// parseSOAPArguments evaluates an HTTP request which is supposed to be a SOAP
// action call and extract the arguments
func (me *Server) parseSOAPArguments(w http.ResponseWriter, r *http.Request, svcID, act string) (args map[string]StateVar, err error) {
	// get request body
	reqBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		me.sendSOAPFault(w,
			SOAPError{
				Code: UPnPErrorHumanRequired,
				Desc: fmt.Sprintf("message body for action '%s' cannot be read", act),
			},
		)
		err = errors.Wrap(err, "request body cannot be read")
		log.Error(err)
		return
	}

	var env soapEnv
	if err = xml.Unmarshal(reqBody, &env); err != nil {
		me.sendSOAPFault(w,
			SOAPError{
				Code: UPnPErrorHumanRequired,
				Desc: fmt.Sprintf("message body for action '%s' cannot be unmarshalled", act),
			},
		)
		err = errors.Wrap(err, "request body cannot be unmarshalled")
		log.Error(err)
		return
	}

	var action soapAct
	if err = xml.Unmarshal(env.Body.Content, &action); err != nil {
		me.sendSOAPFault(w,
			SOAPError{
				Code: UPnPErrorHumanRequired,
				Desc: fmt.Sprintf("message body for action '%s' cannot be unmarshalled", act),
			},
		)
		err = errors.Wrapf(err, "message body for action '%s' cannot be unmarshalled", act)
		log.Error(err)
		return
	}

	args = make(map[string]StateVar)
	for _, arg := range action.Args {
		// get argument spec. Note: Here, we can assume that the maps
		// me.services and me.services[..].actSpecs contain the required
		// elements
		sv, exists := me.services[svcID].actSpecs[act][arg.Name]
		if !exists {
			me.sendSOAPFault(w,
				SOAPError{
					Code: UPnPErrorInvalidArgs,
					Desc: fmt.Sprintf("no specification for argument '%s' of action '%s' found", arg.Name, act),
				},
			)
			err = fmt.Errorf("no specification for argument '%s' of action '%s' found", arg.Name, act)
			log.Error(err)
			return
		}

		// check if argument is valid (i.e. is in the specified range -
		// provided that's a numeric value - or is in the allowed value list -
		// provided it's a string)
		if isValid, errCode := sv.IsValid(arg.Value); !isValid {
			me.sendSOAPFault(w,
				SOAPError{
					Code: errCode,
					Desc: fmt.Sprintf("arg %s is not valid: %s", arg.Name, arg.Value),
				},
			)
			err = fmt.Errorf("arg %s is not valid: %s", arg.Name, arg.Value)
			log.Error(err)
			return
		}

		// create variable for argument and add it to result
		if args[arg.Name], err = newStateVar(sv.Type(), arg.Value); err != nil {
			me.sendSOAPFault(w,
				SOAPError{
					Code: UPnPErrorInvalidArgs,
					Desc: err.Error(),
				},
			)
		}

	}

	return
}

// sendSOAPFault sends a SOAP fault message
func (me *Server) sendSOAPFault(w http.ResponseWriter, soapErr SOAPError) {
	fault := soapErr.marshal()

	setHeader(w, me.ServerString(), len(fault))
	w.WriteHeader(http.StatusInternalServerError)
	n, err := w.Write(fault)
	if err != nil {
		err = errors.Wrap(err, "SOAP fault message cannot be written to HTTP response")
		log.Fatal(err)
	}
	if n < len(fault) {
		err = fmt.Errorf("incomplete write of SOAP fault: %d/%d bytes", n, len(fault))
		log.Fatal(err)
	}
}
