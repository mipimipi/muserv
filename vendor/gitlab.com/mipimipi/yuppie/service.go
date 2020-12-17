package yuppie

import (
	"fmt"
	"strings"

	"gitlab.com/mipimipi/yuppie/desc"
	"gitlab.com/mipimipi/yuppie/internal/events"
)

type (
	// serviceID represents the id of a service
	serviceID string
	// serviceType represents the type of a service
	serviceType string
	// serviceVersion represents the version of a service type
	serviceVersion string
)

// tail returns the pure id part of a service id of the form
// urn:<DOMAIN-NAME>:serviceId:<SERVICE-ID>
func (me serviceID) tail() string {
	s := strings.Split(string(me), ":")
	if len(s) != 4 {
		return ""
	}
	return s[3]
}

// newServiceType creates a new service type
func newServiceType(t string) serviceType {
	s := strings.Split(t, ":")
	if len(s) != 5 {
		err := fmt.Errorf("mal-formed service type: %s", t)
		log.Error(err)
		return serviceType("")
	}
	return serviceType(s[0] + ":" + s[1] + ":" + s[2] + ":" + s[3])
}

// newServiceVersion creates a new service version
func newServiceVersion(t string) (serviceVersion, error) {
	s := strings.Split(t, ":")
	if len(s) != 5 {
		return "", fmt.Errorf("mal-formed service type: %s", t)
	}
	return serviceVersion(s[4]), nil
}

// service returns a UPnP service
type service struct {
	id        serviceID
	typ       serviceType
	ver       serviceVersion
	device    *device
	actSpecs  map[string]map[string](*stateVar)
	stateVars map[string](*stateVar)
	desc      *desc.Service
}

// serviceMap map a service id (only the pure id part) to the corresponding
// service
type serviceMap map[string]*service

// newService creates a new service for a certain id, type and version, based
// on a service description. A listener for multicast eventing is assigned to
// the state variables of the service. It returns a reference to the service.
func newService(id serviceID, typ serviceType, ver serviceVersion, svcDesc *desc.Service, listener func() chan events.StateVar) (*service, error) {
	svc := service{
		id:   id,
		typ:  typ,
		ver:  ver,
		desc: svcDesc,
	}

	// create statevars map
	svc.stateVars = make(map[string](*stateVar))
	for _, sv := range svcDesc.ServiceStateTable {
		var err error
		if svc.stateVars[sv.Name], err = stateVarFromDesc(sv, &svc, listener); err != nil {
			return nil, err
		}
	}

	// create actions map
	svc.actSpecs = make(map[string]map[string](*stateVar))
	for _, act := range svcDesc.Actions {
		if _, exists := svc.actSpecs[act.Name]; exists {
			err := fmt.Errorf("action with name '%s' exists already", act.Name)
			return nil, err
		}

		args := make(map[string](*stateVar))

		svc.actSpecs[act.Name] = args
		// retrieve action arguments
		for _, arg := range act.Arguments {
			// only input arguments are relevant
			if arg.Direction != "in" {
				continue
			}
			// retrieve corresponding state variable
			sv, exists := svc.stateVars[arg.RelatedStateVariable]
			if !exists {
				err := fmt.Errorf("state variable '%s' for argument '%s' not found", arg.RelatedStateVariable, arg.Name)
				return nil, err
			}
			svc.actSpecs[act.Name][arg.Name] = sv
		}
	}

	return &svc, nil
}
