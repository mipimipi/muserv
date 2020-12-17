// Package ssdp implements an SSDP (=Simple Service Discovery Protocol) server
package ssdp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	l "github.com/sirupsen/logrus"
	"gitlab.com/mipimipi/yuppie/internal/network"
	"gitlab.com/mipimipi/yuppie/internal/types"
)

var log *l.Entry

const multicastAddrIPv4 = "239.255.255.250:1900"

var multicastUDPAddr *net.UDPAddr

func init() {
	log = l.WithFields(l.Fields{"srv": "upnp:ssdp"})

	var err error
	multicastUDPAddr, err = net.ResolveUDPAddr("udp4", multicastAddrIPv4)
	if err != nil {
		err = errors.Wrapf(err, "could not resolve address %s", multicastAddrIPv4)
		log.Fatal(err)
	}
}

// Server represents an SSDP server. Such a server is created per network
// interface. If multiple interfaces are available, one SSDP server is created
// per interface
type Server struct {
	bootID   *types.BootID
	configID *types.ConfigID
	inf      net.Interface
	addr     string
	// data from device tree that is relevant for SSDP
	data DiscoveryData
	// index maps keys like device or service type to the corresponding device
	index SearchIndex
	conn  *net.UDPConn
	// channel to trigger stop of notification process
	stopNotify chan struct{}
	// channel to trigger stop response process
	stopResponse chan struct{}
	// channel to receive confirmation about stop of notification process
	notifyStopped chan struct{}
	// channel to receive confirmation about stop of response process
	responseStopped chan struct{}
}

// New creates a new SSDP server
func New(data DiscoveryData, index SearchIndex, bootID *types.BootID, configID *types.ConfigID, inf net.Interface, port int) (srv *Server, err error) {
	log.Tracef("creating SSDP server for interface '%s'", inf.Name)

	srv = new(Server)

	srv.data = data
	srv.index = index
	srv.bootID = bootID
	srv.configID = configID
	srv.inf = inf

	// determine ip4 address of interface
	addrs, err := srv.inf.Addrs()
	if err != nil {
		err = errors.Wrapf(err, "cannot retrieve addresses of interface %s", srv.inf.Name)
		return
	}
	for _, addr := range addrs {
		if addr.(*net.IPNet).IP.To4() != nil {
			a := strings.Split(addr.String(), ":")
			a = strings.Split(a[0], "/")
			srv.addr = a[0]
			break
		}
	}
	if srv.addr == "" {
		err = fmt.Errorf("interface %s has no IP4 address", srv.inf.Name)
		return
	}
	if port != 0 {
		srv.addr += ":" + strconv.Itoa(port)
	}

	return
}

// Connect connects the SSDP server (i.e. starts the notification and search
// response processes)
func (me *Server) Connect() (err error) {
	log.Tracef("connecting SSDP server on interface '%s' ...", me.inf.Name)

	if me.conn, err = network.UDPConn(me.inf, multicastUDPAddr); err != nil {
		err = errors.Wrapf(err, "cannot connect SSDP server on interface %s", me.inf.Name)
		return
	}

	me.stopNotify = make(chan struct{})
	me.stopResponse = make(chan struct{})

	go me.notify()
	go me.listenAndRespond()

	log.Tracef("SSDP server on interface '%s' connected", me.inf.Name)
	return
}

// Disconnect disconnect the SSDP server (i.e. stops the notification and search
// response processes)
func (me *Server) Disconnect(wg *sync.WaitGroup) {
	defer func() {
		close(me.stopNotify)
		close(me.stopResponse)
		wg.Done()
	}()

	log.Tracef("disconnecting SSDP server on interface '%s' ...", me.inf.Name)

	// send stop signals to notify and response loops and wait for stop
	// confirmation
	me.stopNotify <- struct{}{}
	me.stopResponse <- struct{}{}
	<-me.responseStopped
	<-me.notifyStopped

	me.sendByeBye()

	log.Tracef("SSDP server on interface '%s' disconnected", me.inf.Name)
}
