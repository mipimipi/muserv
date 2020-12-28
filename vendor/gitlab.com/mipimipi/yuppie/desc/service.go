package desc

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gitlab.com/mipimipi/go-utils"
)

// Service represents a service description as described in
// https://openconnectivity.org/upnp-specs/UPnP-arch-DeviceArchitecture-v2.0-20200417.pdf
type Service struct {
	XMLName           xml.Name        `xml:"urn:schemas-upnp-org:service-1-0 scpd"`
	ConfigID          uint32          `xml:"configId,attr"`
	SpecVersion       SpecVersion     `xml:"specVersion"`
	Actions           []Action        `xml:"actionList>action"`
	ServiceStateTable []StateVariable `xml:"serviceStateTable>stateVariable"`
}

// ServiceMap maps a service id to the correspondingh service description
type ServiceMap map[string]*Service

// LoadService reads a service description file and creates a service object
// from it
func LoadService(filepath string) (svc *Service, err error) {
	// load root device from file
	f, err := os.Open(filepath)
	if err != nil {
		err = errors.Wrapf(err, "service description file '%s' couldn't be opened", filepath)
		log.Fatal(err)
		return
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		err = errors.Wrapf(err, "service description file '%s' couldn't be read", filepath)
		log.Fatal(err)
		return
	}
	svc = new(Service)
	err = xml.Unmarshal(data, svc)
	if err != nil {
		err = errors.Wrap(err, "service description couldn't be unmarshalled")
		log.Fatal(err)
		return
	}

	return
}

// ClearAttr clear attributes that are under the control of the UPnP server
// such as ConfigID
func (me *Service) ClearAttr() {
	me.ConfigID = 0
}

// Hash calculates the FNV hash of the XML representation of a service description
func (me *Service) Hash() (hash uint64, err error) {
	var s []byte
	if s, err = xml.Marshal(me); err != nil {
		err = errors.Wrap(err, "cannot marshal service")
		log.Fatal(err)
		return
	}

	hash = utils.HashUint64("%x", s)
	return
}

// Validate checks whether the attribute values are OK. Problem messages are
// added to res.
func (me *Service) Validate() (ok bool, res []string) {
	ok = true
	// name space
	if me.XMLName.Space != "urn:schemas-upnp-org:service-1-0" {
		ok = false
		res = append(res, fmt.Sprintf("service: incorrect XML name space: %s", me.XMLName.Space))
	}
	// spec version
	if me.SpecVersion.Minor > me.SpecVersion.Major {
		ok = false
		res = append(res, fmt.Sprintf("service: incorrect spec version: minor=%d major=%d", me.SpecVersion.Minor, me.SpecVersion.Major))
	}
	// action list
	for _, act := range me.Actions {
		b := act.Validate(&res)
		ok = ok && b
	}
	// state variables
	if len(me.ServiceStateTable) == 0 {
		ok = false
		res = append(res, "service: has no state variables")
	}
	for _, v := range me.ServiceStateTable {
		b := v.Validate(&res)
		ok = ok && b
	}

	return
}
