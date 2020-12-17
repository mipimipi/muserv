package yuppie

import (
	"fmt"

	"github.com/pkg/errors"
	"gitlab.com/mipimipi/yuppie/desc"
	"gitlab.com/mipimipi/yuppie/internal/network"
	"gitlab.com/mipimipi/yuppie/internal/ssdp"
)

// createSSDPServers creates SSDP server. One for each suitable network
// interface. If no suitable interface is available, the function returns an#
// error
func (me *Server) createSSDPServers() (err error) {
	log.Trace("creating SSDP servers")

	infs, err := network.Interfaces(me.cfg.Interfaces)
	if err != nil {
		err = errors.Wrap(err, "cannot create SSDP servers")
		log.Fatal(err)
		return
	}

	// create one SSDP server for each interface that is up and is no loopback
	data := me.createDiscoveryData()
	index := me.createSearchIndex()

	for _, inf := range infs {
		ssdp, err := ssdp.New(data, index, me.bootID, me.configID, inf, me.cfg.Port)
		if err != nil {
			err = errors.Wrapf(err, "cannot create SSDP server for interface %s", inf.Name)
			log.Error(err)
			continue
		}
		me.ssdps = append(me.ssdps, ssdp)
	}

	// at least one SSDP server must have been created
	if len(me.ssdps) == 0 {
		err = fmt.Errorf("no SSDP server could be created since no suitable network interfaces was found")
		log.Fatal(err)
		return
	}

	log.Trace("SSDP servers created")
	return
}

// createDiscoveryData creates the data from server that is required by SSDP
// for discovery messages
func (me *Server) createDiscoveryData() (data ssdp.DiscoveryData) {
	data.Server = me.ServerString()
	// {{ADDRESS}} will be replaced by the SSDP server with the address of the
	// network interface of that server
	data.Location = httpProtocol + "://{{ADDRESS}}" + deviceDescPath
	data.MaxAge = me.cfg.MaxAge
	data.AssIDs = getDeviceAssets(&me.Device.Desc.Device, true)

	return
}

// getDeviceAssets extracts the device data that is required for SSDP
func getDeviceAssets(dvc *desc.Device, isRoot bool) (assIDs []ssdp.AssetID) {
	var assID ssdp.AssetID

	// set NTs and USNs for dvc
	if isRoot {
		assID.NT = desc.UPNPRootDeviceType
		assID.USN = dvc.UDN + "::" + desc.UPNPRootDeviceType
		assIDs = append(assIDs, assID)
	}
	assID.NT = dvc.UDN
	assID.USN = dvc.UDN
	assIDs = append(assIDs, assID)
	assID.NT = dvc.DeviceType
	assID.USN = dvc.UDN + "::" + dvc.DeviceType
	assIDs = append(assIDs, assID)

	// set NTs and USNs for service types of dvc
	for _, s := range dvc.Services {
		assID.NT = s.ServiceType
		assID.USN = dvc.UDN + "::" + s.ServiceType
		assIDs = append(assIDs, assID)
	}

	// get NTs and USNs from embedded devices
	for _, d := range dvc.Devices {
		assIDs = append(assIDs, getDeviceAssets(&d, false)...)
	}

	return
}

// createSearchIndex creates an index that maps keys such as device or service
// types to the corresponding device. This data is needed by SSDP for search
// requests
func (me *Server) createSearchIndex() (index ssdp.SearchIndex) {
	index = make(ssdp.SearchIndex)
	getDeviceIndex(&me.Device.Desc.Device, true, &index)
	return
}

// getDeviceIndex creates adds device data to index
func getDeviceIndex(dvc *desc.Device, isRoot bool, index *ssdp.SearchIndex) {
	var ok bool

	if isRoot {
		_, ok = (*index)[desc.UPNPRootDeviceType]
		if !ok {
			(*index)[desc.UPNPRootDeviceType] = new([]string)
			*(*index)[desc.UPNPRootDeviceType] = append(
				*(*index)[desc.UPNPRootDeviceType],
				fmt.Sprintf("%s::%s", dvc.UDN, desc.UPNPRootDeviceType),
			)
		}
	}

	// device uuid
	_, ok = (*index)[dvc.UDN]
	if !ok {
		(*index)[dvc.UDN] = new([]string)
		*(*index)[dvc.UDN] = append(
			*(*index)[dvc.UDN],
			fmt.Sprintf("%s::%s", dvc.UDN, dvc.DeviceType),
		)
	}

	// device type
	_, ok = (*index)[dvc.DeviceType]
	if !ok {
		(*index)[dvc.DeviceType] = new([]string)
	}
	*(*index)[dvc.DeviceType] = append(
		*(*index)[dvc.DeviceType],
		fmt.Sprintf("%s::%s", dvc.UDN, dvc.DeviceType),
	)

	// service type
	for _, s := range dvc.Services {
		_, ok = (*index)[s.ServiceType]
		if !ok {
			(*index)[s.ServiceType] = new([]string)
		}
		*(*index)[s.ServiceType] = append(
			*(*index)[s.ServiceType],
			fmt.Sprintf("%s::%s", dvc.UDN, s.ServiceType),
		)
	}

	// get keys from embedded devices
	for _, d := range dvc.Devices {
		getDeviceIndex(&d, false, index)
	}
}
