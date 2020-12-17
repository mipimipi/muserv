// Package network contains function of facilitate sending and receiving
// messages via UDP and TCP
package network

import (
	"net"

	"github.com/pkg/errors"
	l "github.com/sirupsen/logrus"
)

var log *l.Entry

func init() {
	// logging
	log = l.WithFields(l.Fields{"srv": "upnp:network"})
}

// UDPMsgRepetitions is the number of repetitive transmission of UDP
// messages. According to the UPnP Device Architecture 2.0 spec, one
// message can be sent up to 3 times
const UDPMsgRepetitions = 3

// Interfaces returns the network interfaces that are available on the machine
// (i.e. the interfaces that are up and that are no loopback). If wanted is not
// empty, the content of that array is interpreted as interface names and only
// these for these names the corresponding network interfaces are determined
// and returned.
func Interfaces(wanted []string) (infs []net.Interface, err error) {
	// if interfaces have been configured: take them
	// otherwise use all interfaces of this machine
	var inf0s []net.Interface
	if len(wanted) > 0 {
		for _, name := range wanted {
			inf, err := net.InterfaceByName(name)
			if err != nil {
				log.Errorf("cannot determine interface '%s': %v", name, err)
				continue
			}
			inf0s = append(infs, *inf)
		}
	} else {
		log.Trace("get network interfaces of that machine")

		if inf0s, err = net.Interfaces(); err != nil {
			err = errors.Wrap(err, "cannot determine interfaces")
			return
		}
		log.Tracef("found %d interfaces", len(inf0s))
	}

	// collect interfaces that
	// (1) are up
	// (2) are no loopback interface
	// (3) have an ip4 address
	for _, inf := range inf0s {
		if inf.Flags&net.FlagUp == 0 || inf.Flags&net.FlagLoopback != 0 || inf.MTU <= 0 {
			continue
		}

		// get addresses of inf
		addrs, err := inf.Addrs()
		if err != nil {
			log.Errorf("cannot determine IP addresses of interface %s", inf.Name)
			continue
		}
		// check if one of these addresses is an ip4 address
		for _, addr := range addrs { // get ipv4 address
			ip4Addr := addr.(*net.IPNet).IP.To4()
			if ip4Addr != nil {
				infs = append(infs, inf)
				break
			}
		}
	}

	return
}
