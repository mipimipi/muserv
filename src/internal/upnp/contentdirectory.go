package upnp

// this file contains the handler functions for the actions of the content
// directory service

import (
	"fmt"

	"github.com/pkg/errors"
	"gitlab.com/mipimipi/muserv/src/internal/content"
	"gitlab.com/mipimipi/yuppie"
)

// names of arguments of the browse action of the content directory service
const (
	browseReqArgObjID     = "ObjectID"
	browseReqArgMode      = "BrowseFlag"
	browseReqArgCount     = "RequestedCount"
	browseReqArgStart     = "StartingIndex"
	browseRespArgResult   = "Result"
	browseRespArgReturned = "NumberReturned"
	browseRespArgTotal    = "TotalMatches"
	browseRespArgUpdateID = "UpdateID"
)

// handler for action Browse()
func (me *Server) browse(reqArgs map[string]yuppie.StateVar) (respArgs yuppie.SOAPRespArgs, soapErr yuppie.SOAPError) {
	// retrieve and check input arguments
	if len(reqArgs) == 0 {
		log.Error("no arguments passed to Browse action")
		soapErr = yuppie.SOAPError{
			Code: yuppie.UPnPErrorInvalidArgs,
			Desc: "no arguments passed to Browse action",
		}
		return
	}
	for name, value := range reqArgs {
		log.Tracef("arg %s=%s", name, value.String())
	}
	objID, exists := reqArgs[browseReqArgObjID]
	var (
		err error
		id  content.ObjID
	)
	if exists {
		id, err = content.ObjIDFromString(objID.String())
	}
	if !exists || err != nil {
		log.Errorf("invalid ObjectID argument in browse action: '%s'", objID.String())
		soapErr = yuppie.SOAPError{
			Code: yuppie.UPnPErrorInvalidArgs,
			Desc: fmt.Sprintf("invalid ObjectID argument in browse action: '%s'", objID.String()),
		}
		return
	}
	mode, exists := reqArgs[browseReqArgMode]
	if !exists || (mode.String() != content.ModeChildren && mode.String() != content.ModeMetadata) {
		log.Errorf("invalid BrowseFlag argument in browse action: '%d'", id)
		soapErr = yuppie.SOAPError{
			Code: yuppie.UPnPErrorInvalidArgs,
			Desc: fmt.Sprintf("invalid BrowseFlag argument in browse action: '%d'", id),
		}
		return
	}
	var start, wanted uint32
	soapVar, exists := reqArgs[browseReqArgStart]
	if exists {
		start = soapVar.Get().(uint32)
	}
	soapVar, exists = reqArgs[browseReqArgCount]
	if exists {
		wanted = soapVar.Get().(uint32)
	}

	// execute browse
	result, returned, total, err := me.cnt.Browse(
		id,
		mode.String(),
		start,
		wanted,
	)
	if err != nil {
		soapErr = yuppie.SOAPError{
			Code: yuppie.UPnPErrorActionFailed,
			Desc: "error when browsing the music",
		}
		log.Error(errors.Wrap(err, "error when browsing the music"))
		return
	}

	// create output arguments
	updateID, _ := me.StateVariable(svcIDContDir, svSystemUpdateID)
	respArgs = yuppie.SOAPRespArgs{
		browseRespArgResult:   result,
		browseRespArgReturned: fmt.Sprintf("%d", returned),
		browseRespArgTotal:    fmt.Sprintf("%d", total),
		browseRespArgUpdateID: updateID.String(),
	}

	return
}

// handler for action GetSearchCapabilities()
func (me *Server) getSearchCapabilities(reqArgs map[string]yuppie.StateVar) (respArgs yuppie.SOAPRespArgs, soapErr yuppie.SOAPError) {
	sv, exists := me.StateVariable(svcIDContDir, svSearchCapabilities)
	if !exists {
		soapErr = yuppie.SOAPError{
			Code: yuppie.UPnPErrorActionFailed,
			Desc: fmt.Sprintf("state variable '%s' could not be retrieved", svSearchCapabilities),
		}
		return
	}

	respArgs = yuppie.SOAPRespArgs{"SearchCaps": sv.String()}
	return
}

// handler for action GetSortCapabilities()
func (me *Server) getSortCapabilities(reqArgs map[string]yuppie.StateVar) (respArgs yuppie.SOAPRespArgs, soapErr yuppie.SOAPError) {
	sv, exists := me.StateVariable(svcIDContDir, svSortCapabilities)
	if !exists {
		soapErr = yuppie.SOAPError{
			Code: yuppie.UPnPErrorActionFailed,
			Desc: fmt.Sprintf("state variable '%s' could not be retrieved", svSortCapabilities),
		}
		return
	}

	respArgs = yuppie.SOAPRespArgs{"SortCaps": sv.String()}
	return
}

// handler for action GetFeatureList()
func (me *Server) getFeatureList(reqArgs map[string]yuppie.StateVar) (respArgs yuppie.SOAPRespArgs, soapErr yuppie.SOAPError) {
	sv, exists := me.StateVariable(svcIDContDir, svFeatureList)
	if !exists {
		soapErr = yuppie.SOAPError{
			Code: yuppie.UPnPErrorActionFailed,
			Desc: fmt.Sprintf("state variable '%s' could not be retrieved", svFeatureList),
		}
		return
	}

	respArgs = yuppie.SOAPRespArgs{"FeatureList": sv.String()}
	return
}

// handler for action GetSystemUpdateID()
func (me *Server) getSystemUpdateID(reqArgs map[string]yuppie.StateVar) (respArgs yuppie.SOAPRespArgs, soapErr yuppie.SOAPError) {
	sv, exists := me.StateVariable(svcIDContDir, svSystemUpdateID)
	if !exists {
		soapErr = yuppie.SOAPError{
			Code: yuppie.UPnPErrorActionFailed,
			Desc: fmt.Sprintf("state variable '%s' could not be retrieved", svSystemUpdateID),
		}
		return
	}

	respArgs = yuppie.SOAPRespArgs{"Id": sv.String()}
	return
}

// handler for action GetServiceResetToken()
func (me *Server) getServiceResetToken(reqArgs map[string]yuppie.StateVar) (respArgs yuppie.SOAPRespArgs, soapErr yuppie.SOAPError) {
	sv, exists := me.StateVariable(svcIDContDir, svServiceResetToken)
	if !exists {
		soapErr = yuppie.SOAPError{
			Code: yuppie.UPnPErrorActionFailed,
			Desc: fmt.Sprintf("state variable '%s' could not be retrieved", svServiceResetToken),
		}
		return
	}

	respArgs = yuppie.SOAPRespArgs{"ResetToken": sv.String()}
	return
}
