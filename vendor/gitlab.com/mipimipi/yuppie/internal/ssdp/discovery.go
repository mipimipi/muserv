package ssdp

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/fwojciec/clock"
	utils "gitlab.com/mipimipi/go-utils"
	"gitlab.com/mipimipi/yuppie/internal/network"
)

// AssetID contains the NT and USN fields of devices and services
type AssetID struct{ NT, USN string }

// DiscoveryData contains the data from a device tree that is required for
// alive or byebye notifications
type DiscoveryData struct {
	Location string
	AssIDs   []AssetID
	Server   string
	MaxAge   int
}

// String returns a string representation of DiscoveryData
func (me *DiscoveryData) String() (s string) {
	s = fmt.Sprintf("Location: %s\n", me.Location)
	s += fmt.Sprintf("Server: %s\n", me.Server)
	s += fmt.Sprintf("MaxAge: %d\n", me.MaxAge)
	for _, assID := range me.AssIDs {
		s += fmt.Sprintf("\tNT: %s\n", assID.NT)
		s += fmt.Sprintf("\tUSN: %s\n", assID.USN)
	}

	return
}

// notify sends alive messages regularly
func (me *Server) notify() {
	utils.RandomNap(1000)
	me.sendAlive()
	me.notifyStopped = make(chan struct{})

	// define ticker. Per UPnP Device Architecture 2.0 spec, the alive
	// notifications shall be sent "at a randomly-distributed interval of less
	// than one half of the advertisement expiration time"
	ticker := clock.NewRandomTicker(0, time.Duration(me.data.MaxAge)*time.Second/2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			me.sendAlive()
		case <-me.stopNotify:
			close(me.notifyStopped)
			log.Tracef("notify stopped on interface '%s'", me.inf.Name)
			return
		}
	}
}

// sendAlive sends an alive message
func (me *Server) sendAlive() {
	// send alive messages 3 times as required by the UPnP Device Architecture 2.0
	for i := 0; i < network.UDPMsgRepetitions; i++ {
		// sleep for a few hundert milliseconds as required by the UPnP Device
		// Architecture 2.0
		utils.RandomNap(1000)
		// send alive messages
		for _, assID := range me.data.AssIDs {
			msg := new(bytes.Buffer)
			fmt.Fprint(msg, "NOTIFY * HTTP/1.1\r\n")
			fmt.Fprintf(msg, "HOST: %s\r\n", multicastAddrIPv4)
			fmt.Fprintf(msg, "NT: %s\r\n", assID.NT)
			fmt.Fprintf(msg, "NTS: %s\r\n", "ssdp:alive")
			fmt.Fprintf(msg, "USN: %s\r\n", assID.USN)
			fmt.Fprintf(msg, "LOCATION: %s\r\n", strings.Replace(me.data.Location, "{{ADDRESS}}", me.addr, -1))
			fmt.Fprintf(msg, "CACHE-CONTROL: max-age=%d\r\n", me.data.MaxAge)
			fmt.Fprintf(msg, "BOOTID.UPNP.ORG: %d\r\n", me.bootID.Val())
			fmt.Fprintf(msg, "CONFIG.UPNP.ORG: %d\r\n", me.configID.Val())
			// add empty row at the end as required by the UPnP Device
			// Architecture 2.0
			fmt.Fprint(msg, "\r\n")

			if err := network.SendUDP(me.conn, multicastUDPAddr, msg.Bytes()); err != nil {
				continue
			}
		}
	}
	log.Tracef("sent alive messages on interface '%s'", me.inf.Name)
}

// sendByeBye send a byebye message
func (me *Server) sendByeBye() {
	// send alive messages 3 times as required by the UPnP Device Architecture 2.0
	for i := 0; i < network.UDPMsgRepetitions; i++ {
		// sleep for a few hundert milliseconds as required by the UPnP Device
		// Architecture 2.0
		utils.RandomNap(1000)
		// send byebye messages
		for _, assID := range me.data.AssIDs {
			msg := new(bytes.Buffer)
			fmt.Fprint(msg, "NOTIFY * HTTP/1.1\r\n")
			fmt.Fprintf(msg, "HOST: %s\r\n", multicastAddrIPv4)
			fmt.Fprintf(msg, "NT: %s\r\n", assID.NT)
			fmt.Fprintf(msg, "NTS: %s\r\n", "ssdp:byebye")
			fmt.Fprintf(msg, "USN: %s\r\n", assID.USN)
			fmt.Fprintf(msg, "BOOTID.UPNP.ORG: %d\r\n", me.bootID.Val())
			fmt.Fprintf(msg, "CONFIG.UPNP.ORG: %d\r\n", me.configID.Val())
			// add empty row at the end as required by the UPnP Device
			// Architecture 2.0
			fmt.Fprint(msg, "\r\n")

			if err := network.SendUDP(me.conn, multicastUDPAddr, msg.Bytes()); err != nil {
				continue
			}
		}
	}
	log.Tracef("sent byebye messages on interface '%s'", me.inf.Name)
}
