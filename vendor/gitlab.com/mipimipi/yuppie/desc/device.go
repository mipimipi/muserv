// Package desc implements data types to map to the content from description
// XML files
package desc

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	l "github.com/sirupsen/logrus"
	"gitlab.com/mipimipi/go-utils"
)

var log *l.Entry

func init() {
	log = l.WithFields(l.Fields{"srv": "upnp:types"})
}

const (
	// UPNPRootDeviceType is the UPnP root device indicator
	UPNPRootDeviceType = "upnp:rootdevice"
)

var reDvcType = regexp.MustCompile(`urn:.+:device:.*:\d+`)

// SpecVersion is part of a RootDevice, describes the version of the
// specification that the data adheres to.
type SpecVersion struct {
	Major int32 `xml:"major"`
	Minor int32 `xml:"minor"`
}

// RootDevice represents a device description as described in
// https://openconnectivity.org/upnp-specs/UPnP-arch-DeviceArchitecture-v2.0-20200417.pdf
type RootDevice struct {
	XMLName     xml.Name    `xml:"urn:schemas-upnp-org:device-1-0 root"`
	ConfigID    uint32      `xml:"configId,attr"`
	SpecVersion SpecVersion `xml:"specVersion"`
	Device      Device      `xml:"device"`
}

// Device represents a device description
type Device struct {
	DeviceType       string             `xml:"deviceType"`
	FriendlyName     string             `xml:"friendlyName"`
	Manufacturer     string             `xml:"manufacturer"`
	ManufacturerURL  string             `xml:"manufacturerURL"`
	ModelDescription string             `xml:"modelDescription"`
	ModelName        string             `xml:"modelName"`
	ModelNumber      string             `xml:"modelNumber"`
	ModelURL         string             `xml:"modelURL"`
	SerialNumber     string             `xml:"serialNumber"`
	UDN              string             `xml:"UDN"`
	UPC              string             `xml:"UPC,omitempty"`
	Icons            []Icon             `xml:"iconList>icon,omitempty"`
	Services         []ServiceReference `xml:"serviceList>service,omitempty"`
	Devices          []Device           `xml:"deviceList>device,omitempty"`
	PresentationURL  string             `xml:"presentationURL"`
}

// Icon represents the icon part of a device description
type Icon struct {
	Mimetype string `xml:"mimetype"`
	Width    uint32 `xml:"width"`
	Height   uint32 `xml:"height"`
	Depth    uint32 `xml:"depth"`
	URL      string `xml:"url"`
}

// LoadRootDevice reads a root device description file and creates a root
// device object from it
func LoadRootDevice(filepath string) (dvc *RootDevice, err error) {
	// load root device from file
	f, err := os.Open(filepath)
	if err != nil {
		err = errors.Wrapf(err, "device description file '%s' couldn't be opened", filepath)
		log.Fatal(err)
		return
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		err = errors.Wrapf(err, "device description file '%s' couldn't be read", filepath)
		log.Fatal(err)
		return
	}
	dvc = new(RootDevice)
	err = xml.Unmarshal(data, dvc)
	if err != nil {
		err = errors.Wrap(err, "device description couldn't be unmarshalled")
		log.Fatal(err)
		return
	}

	return
}

// ClearAttr clear attributes that are under the control of the UPnP server
// such as ConfigID
func (me *RootDevice) ClearAttr() {
	me.ConfigID = 0
	me.Device.ClearAttr()
}

// Hash calculates the FNV hash of the XML representation of a root device description
func (me *RootDevice) Hash() (hash uint64, err error) {
	var s []byte
	if s, err = xml.Marshal(me); err != nil {
		err = errors.Wrap(err, "cannot marshal root device")
		log.Fatal(err)
		return
	}

	hash = utils.HashUint64("%x", s)
	return
}

// Validate checks whether the attribute values are OK. Problem messages are
// stored in res.
func (me *RootDevice) Validate() (ok bool, res []string) {
	ok = true
	// name space
	if me.XMLName.Space != "urn:schemas-upnp-org:device-1-0" {
		ok = false
		res = append(res, fmt.Sprintf("root device: incorrect XML name space: %s", me.XMLName.Space))
	}
	// spec version
	if me.SpecVersion.Minor > me.SpecVersion.Major {
		ok = false
		res = append(res, fmt.Sprintf("root device: incorrect spec version: minor=%d major=%d", me.SpecVersion.Minor, me.SpecVersion.Major))
	}

	// validate device
	_ = me.Device.Validate(&res)

	return
}

// ClearAttr clear attributes that are under the control of the UPnP server
func (me *Device) ClearAttr() {
	// clear attributes of service references
	svcRefs := &(me.Services)
	for i := 0; i < len(me.Services); i++ {
		(*svcRefs)[i].ClearAttr()
	}
	// clear attributes of embedded devices
	dvcs := &(me.Devices)
	for i := 0; i < len(me.Devices); i++ {
		(*dvcs)[i].ClearAttr()
	}
}

// Trim removes leading and trailing spaces for some attributes
func (me *Device) Trim() {
	me.DeviceType = strings.TrimSpace(me.DeviceType)
	me.ManufacturerURL = strings.TrimSpace(me.ManufacturerURL)
	me.ModelURL = strings.TrimSpace(me.ModelURL)
	me.UDN = strings.TrimSpace(me.UDN)
	me.UPC = strings.TrimSpace(me.UPC)
	me.PresentationURL = strings.TrimSpace(me.DeviceType)
	me.DeviceType = strings.TrimSpace(me.PresentationURL)
}

// Validate checks whether the attribute values are OK. Problem messages are
// added to res.
func (me Device) Validate(res *[]string) (ok bool) {
	me.Trim()

	ok = true
	// DeviceType
	if !reDvcType.MatchString(me.DeviceType) {
		ok = false
		*res = append(*res, fmt.Sprintf("device: wrong device type: %s", me.DeviceType))
	}
	// friendly name
	if me.FriendlyName == "" {
		ok = false
		*res = append(*res, "device: friendly name must not be empty")
	}
	// manufacturer
	if me.Manufacturer == "" {
		ok = false
		*res = append(*res, "device: manufacturer must not be empty")
	}
	// manufacturer URL
	if me.ManufacturerURL != "" {
		if _, err := url.ParseRequestURI(me.ManufacturerURL); err != nil {
			ok = false
			*res = append(*res, fmt.Sprintf("device: incorrect manufacturer URL: %s", me.ManufacturerURL))
		}
	}
	// model name
	if me.ModelName == "" {
		ok = false
		*res = append(*res, "device: model name must not be empty")
	}
	// model URL
	if me.ModelURL != "" {
		if _, err := url.ParseRequestURI(me.ModelURL); err != nil {
			ok = false
			*res = append(*res, fmt.Sprintf("device: incorrect manufacturer URL: %s", me.ManufacturerURL))
		}
	}
	// UDN
	// note: uuid.Parse expects uuid field to start with "urn:", but the UPnP
	// spec doesn't require that a UDN field starts with that
	if _, err := uuid.Parse("urn:" + me.UDN); err != nil {
		ok = false
		*res = append(*res, fmt.Sprintf("device: incorrect UDN '%s': %v", me.UDN, err.Error()))
	}
	// presentation URL
	if me.PresentationURL != "" {
		if _, err := url.Parse(me.PresentationURL); err != nil {
			ok = false
			*res = append(*res, fmt.Sprintf("device: incorrect presentation URL: %s", me.ManufacturerURL))
		}
	}

	// validate service list
	for _, svcRef := range me.Services {
		b := svcRef.Validate(res)
		ok = ok && b
	}

	// validate embedded devices
	for _, dvc := range me.Devices {
		b := dvc.Validate(res)
		ok = ok && b
	}

	return
}
