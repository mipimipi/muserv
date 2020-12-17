package desc

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	reSvcType = regexp.MustCompile(`urn:.+:service:.*:\d+`)
	reSvcID   = regexp.MustCompile(`urn:.+:serviceId:.*`)
)

// ServiceReference represents the service information that is contained in a
// device description
type ServiceReference struct {
	ServiceType string `xml:"serviceType"`
	ServiceID   string `xml:"serviceId"`
	SCPDURL     string `xml:"SCPDURL"`
	ControlURL  string `xml:"controlURL"`
	EventSubURL string `xml:"eventSubURL"`
}

// Trim removes leading and trailing spaces for some attributes
func (me *ServiceReference) Trim() {
	me.ServiceType = strings.TrimSpace(me.ServiceType)
	me.ServiceID = strings.TrimSpace(me.ServiceID)
	me.SCPDURL = strings.TrimSpace(me.SCPDURL)
	me.ControlURL = strings.TrimSpace(me.ControlURL)
	me.EventSubURL = strings.TrimSpace(me.EventSubURL)
}

// Validate executes a trim and checks whether the attribute values are OK.
// Problem messages are added to res.
func (me *ServiceReference) Validate(res *[]string) (ok bool) {
	me.Trim()

	ok = true
	// Service Type
	if !reSvcType.MatchString(me.ServiceType) {
		ok = false
		*res = append(*res, "service ref: wrong service type: "+me.ServiceType)
	}
	// Service ID
	if !reSvcID.MatchString(me.ServiceID) {
		ok = false
		*res = append(*res, "service ref: mal-formed service ID: "+me.ServiceID)
	}
	// SCPDURL
	if _, err := url.Parse(me.SCPDURL); err != nil {
		ok = false
		*res = append(*res, "service ref: incorrect SCPDURL: "+me.SCPDURL)
	}
	// control URL
	if _, err := url.Parse(me.ControlURL); err != nil {
		ok = false
		*res = append(*res, "service ref: incorrect control URL: "+me.ControlURL)
	}
	// event URL
	if _, err := url.Parse(me.EventSubURL); err != nil {
		ok = false
		*res = append(*res, "service ref: event subscription URL: "+me.EventSubURL)
	}

	return
}
