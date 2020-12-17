package yuppie

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.com/mipimipi/yuppie/desc"
	"gitlab.com/mipimipi/yuppie/internal/events"
)

// rootDevice represents an UPnP root device
type rootDevice struct {
	Desc *desc.RootDevice
	*device
}

// device represents an UPnP device
type device struct {
	UDN      string
	services []*service
	devices  []*device
}

// createFromDesc take devices and service descriptions and creates device and
// service objects from it. listener is a listener for multicast eventing. It
// returns a reference to the root device and a map of services.
// Note: The key of svcDesc must correspong to the service ids in dvcDesc to
// allow a proper connection between of services and devides
func createFromDesc(dvcDesc *desc.RootDevice, svcDescs desc.ServiceMap, listener func() chan events.StateVar) (*rootDevice, serviceMap, error) {
	svcs := make(serviceMap)

	dvc, err := newDevice(&dvcDesc.Device, svcDescs, listener, svcs)
	if err != nil {
		err = errors.Wrap(err, "cannot create device from device description")
		log.Fatal(err)
		return nil, nil, err
	}

	root := rootDevice{
		dvcDesc,
		dvc,
	}

	return &root, svcs, nil
}

// newDevice creates a new device. For embedded devices, it's called
// recursively. See also CreateFromDesc.
func newDevice(dvcDesc *desc.Device, svcDescs desc.ServiceMap, listener func() chan events.StateVar, svcs serviceMap) (dvc *device, err error) {
	log.Tracef("creating new device ...")

	dvc = new(device)

	// set UDN from description. If the UDN is empty or contains and empty
	// UUID, generate a new UUID
	if len(dvcDesc.UDN) == 0 || strings.ToLower(dvcDesc.UDN) == "uuid:" {
		dvc.UDN = "uuid:" + uuid.New().String()
	} else {
		dvc.UDN = dvcDesc.UDN
	}

	// process service info
	// note: this for-loop-variant had to be chosen since
	//       "for ... range ..." copies the array items which does not
	//       allow to change array items
	svcRefs := &(dvcDesc.Services)
	for i := 0; i < len(*svcRefs); i++ {
		id := serviceID((*svcRefs)[i].ServiceID)
		typ := newServiceType((*svcRefs)[i].ServiceType)

		var ver serviceVersion
		ver, err = newServiceVersion((*svcRefs)[i].ServiceType)
		if err != nil {
			err = errors.Wrapf(err, "could not determine service version for device '%s'", dvc.UDN)
			return
		}

		if _, exists := svcs[id.tail()]; exists {
			err = fmt.Errorf("service id '%s' is used multiple times", id.tail())
			return
		}

		svcDesc, exists := svcDescs[id.tail()]
		if !exists {
			err = fmt.Errorf("no descrption for service id '%s'", id.tail())
			return
		}

		svc, err := newService(id, typ, ver, svcDesc, listener)
		if err != nil {
			err = errors.Wrapf(err, "cannot create service of id '%s'", id)
			return nil, err
		}
		svc.device = dvc

		dvc.services = append(dvc.services, svc)
		svcs[id.tail()] = svc
	}

	// recursion: embedded devices
	for _, subDesc := range dvcDesc.Devices {
		subDvc, err := newDevice(&subDesc, svcDescs, listener, svcs)
		if err != nil {
			err = errors.Wrapf(err, "cannot create sub device of device '%s", dvc.UDN)
			return nil, err
		}
		dvc.devices = append(dvc.devices, subDvc)
	}

	log.Tracef("new device created")

	return
}
