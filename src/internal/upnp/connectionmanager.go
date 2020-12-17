package upnp

// this file contains the handler functions for the actions of the connection
// manager service

import (
	"fmt"

	"gitlab.com/mipimipi/yuppie"
)

// name of protocol info source argument of current connection info action of
// the connection manager service
const protocolInfoSource = "Source"

// handler for action GetProtocolInfo()
func (me *Server) getProtocolInfo(reqArgs map[string]yuppie.StateVar) (respArgs yuppie.SOAPRespArgs, soapErr yuppie.SOAPError) {
	sv, exists := me.StateVariable(svcIDConnMgr, svSourceProtocolInfo)
	if !exists {
		soapErr = yuppie.SOAPError{
			Code: yuppie.UPnPErrorActionFailed,
			Desc: fmt.Sprintf("state variable '%s' could not be retrieved", svSourceProtocolInfo),
		}
		return
	}

	// since muserv only "outputs" data. Sink must be empty as required by
	// ConnectionManager:2, Service Template Version 1.01
	respArgs = yuppie.SOAPRespArgs{
		"Source": sv.String(),
		"Sink":   "",
	}

	return
}

// handler for action GetCurrentConnectionIDs()
func (me *Server) getCurrentConnectionIDs(reqArgs map[string]yuppie.StateVar) (respArgs yuppie.SOAPRespArgs, soapErr yuppie.SOAPError) {
	sv, exists := me.StateVariable(svcIDConnMgr, svCurrentConnectionIDs)
	if !exists {
		soapErr = yuppie.SOAPError{
			Code: yuppie.UPnPErrorActionFailed,
			Desc: fmt.Sprintf("state variable '%s' could not be retrieved", svCurrentConnectionIDs),
		}
		return
	}

	respArgs = yuppie.SOAPRespArgs{"ConnectionIDs": sv.String()}
	return
}

// handler for action GetCurrentConnectionInfo()
func (me *Server) getCurrentConnectionInfo(reqArgs map[string]yuppie.StateVar) (respArgs yuppie.SOAPRespArgs, soapErr yuppie.SOAPError) {
	// since muserv does not implement the action PrepareForConnection(), the
	// action can only respond connection ID 0 as required by
	// ConnectionManager:2, Service Template Version 1.01
	src, exists := reqArgs[protocolInfoSource]
	if len(reqArgs) != 1 || !exists || src.String() != "0" {
		soapErr = yuppie.SOAPError{
			Code: 706,
			Desc: "the connection reference argument does not refer to a valid connection established by this service",
		}
		return
	}

	// get state variable SourceProtocolInfo
	sv, exists := me.StateVariable(svcIDConnMgr, svSourceProtocolInfo)
	if !exists {
		soapErr = yuppie.SOAPError{
			Code: yuppie.UPnPErrorActionFailed,
			Desc: fmt.Sprintf("state variable '%s' could not be retrieved", svSourceProtocolInfo),
		}
		return
	}

	// since muserv does not implement the action PrepareForConnection(), the
	// action can only respond a limited set of information as required by
	// ConnectionManager:2, Service Template Version 1.01
	respArgs = yuppie.SOAPRespArgs{
		"RcsID":                 "0",
		"AVTransportID":         "0",
		"ProtocolInfo":          sv.String(),
		"PeerConnectionManager": "",
		"PeerConnectionID":      "-1",
		"Direction":             "Output",
		"Status":                "OK",
	}

	return
}
