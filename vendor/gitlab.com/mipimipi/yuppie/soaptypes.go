package yuppie

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"time"
)

// SOAPRespArgs maps argument name to argument value
type SOAPRespArgs map[string]string

// soapEnv represents a SOAP envelope
type soapEnv struct {
	XMLName       xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	EncodingStyle string   `xml:"encodingStyle,attr"`
	Body          soapBody `xml:"Body"`
}

// soapBody represents a SOAP body
type soapBody struct {
	Content []byte `xml:",innerxml"`
}

// marshal marshals an XML snippet for a SOAP Body
func (me soapEnv) marshal() []byte {
	buf := new(bytes.Buffer)

	fmt.Fprint(buf, "<?xml version=\"1.0\" encoding=\"utf-8\"?>")
	fmt.Fprint(buf, "<s:Envelope xmlns:s=\"http://schemas.xmlsoap.org/soap/envelope/\" s:encodingStyle=\"http://schemas.xmlsoap.org/soap/encoding/\">")
	fmt.Fprint(buf, "<s:Body>")
	_, _ = buf.Write(me.Body.Content)
	fmt.Fprint(buf, "</s:Body>")
	fmt.Fprint(buf, "</s:Envelope>")

	return buf.Bytes()
}

// soapAct represents a SOAP action
type soapAct struct {
	Args []soapArg `xml:",any"`
}

// soapActionResponse represents a response to a SOAP action
type soapActResp struct {
	serviceType string
	serviceVer  string
	name        string
	args        SOAPRespArgs
}

// marshal marshals an XML snippet for an response to a SOAP action
func (me soapActResp) marshal() []byte {
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "<u:%sResponse xmlns:u=\"%s:%s\">", me.name, me.serviceType, me.serviceVer)
	for name, value := range me.args {
		fmt.Fprintf(buf, "<%s>%s</%s>", name, html.EscapeString(value), name)
	}
	fmt.Fprintf(buf, "</u:%sResponse>", me.name)

	return soapEnv{Body: soapBody{Content: buf.Bytes()}}.marshal()
}

// soapArg represents an argument of a SOAP action
type soapArg struct {
	Name  string
	Value string
}

// UnmarshalXML unmarshals an argument of a SOAP action
func (me *soapArg) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	me.Name = start.Name.Local
	return d.DecodeElement(&me.Value, &start)
}

// UPnPErrorCode represents an UPnP error code
type UPnPErrorCode uint

const (
	// UPnPErrorInvalidAction is the code foran invalid action
	UPnPErrorInvalidAction UPnPErrorCode = 400
	// UPnPErrorInvalidArgs is the code for invalid arguments
	UPnPErrorInvalidArgs UPnPErrorCode = 402
	// UPnPErrorActionFailed is the code for a failed action
	UPnPErrorActionFailed UPnPErrorCode = 501
	// UPnPErrorArgValInvalid is the code for an invalid argument value
	UPnPErrorArgValInvalid UPnPErrorCode = 600
	// UPnPErrorArgValOutOfRange is the code for an argument value that is out
	// of range
	UPnPErrorArgValOutOfRange UPnPErrorCode = 601
	// UPnPErrorOptActionNotImplemented is the code for an action that is
	// called but not implemented
	UPnPErrorOptActionNotImplemented UPnPErrorCode = 602
	// UPnPErrorHumanRequired indicates that human interaction is required
	UPnPErrorHumanRequired UPnPErrorCode = 604
	// UPnPErrorStrTooLong indicates that a string is too long
	UPnPErrorStrTooLong UPnPErrorCode = 605
)

// SOAPError represents a SOAP error
type SOAPError struct {
	Code UPnPErrorCode
	Desc string
}

// IsNil returns true if the SOAP error is not really an error
func (me SOAPError) IsNil() bool {
	return me.Code == 0 && me.Desc == ""
}

// marshal marshals an XML snippet for a SOAP error
func (me SOAPError) marshal() []byte {
	buf := new(bytes.Buffer)
	fmt.Fprint(buf, "<s:Fault>")
	fmt.Fprint(buf, "<faultcode>s:Client</faultcode>")
	fmt.Fprint(buf, "<faultstring>UPnPError</faultstring>")
	fmt.Fprint(buf, "<detail>")
	fmt.Fprint(buf, "<UPnPError xmlns=\"urn:schemas-upnp-org:control-1-0\">")
	fmt.Fprintf(buf, "<errorCode>%d</errorCode>", me.Code)
	fmt.Fprintf(buf, "<errorDescription>%s</errorDescription>", html.EscapeString(me.Desc))
	fmt.Fprint(buf, "</UPnPError>")
	fmt.Fprint(buf, "</detail>")
	fmt.Fprint(buf, "</s:Fault>")

	return soapEnv{Body: soapBody{Content: buf.Bytes()}}.marshal()
}

// timeOfDay is used in cases where SOAP "time" or "time.tz" is used. This type
// definition is copied from https://github.com/huin/goupnp, copyright 2013
// John Beisley <johnbeisleyuk@gmail.com>.
type timeOfDay struct {
	// Duration of time since midnight.
	FromMidnight time.Duration

	// Set to true if Offset is specified. If false, then the timezone is
	// unspecified (and by ISO8601 - implies some "local" time).
	HasOffset bool

	// Offset is non-zero only if time.tz is used. It is otherwise ignored. If
	// non-zero, then it is regarded as a UTC offset in seconds. Note that the
	// sub-minutes is ignored by the marshal function.
	Offset int
}
