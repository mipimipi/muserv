package network

import (
	"fmt"
	"net"

	"github.com/pkg/errors"
	"golang.org/x/net/ipv4"
)

// UDPConn creates a UDP network connection via interface inf to address addr
func UDPConn(inf net.Interface, addr *net.UDPAddr) (conn *net.UDPConn, err error) {
	if conn, err = net.ListenMulticastUDP("udp", &inf, addr); err != nil {
		err = errors.Wrapf(err, "cannot listen to multicast UDP on interface %s", inf.Name)
		return
	}
	p := ipv4.NewPacketConn(conn)
	if err = p.SetMulticastTTL(2); err != nil {
		err = errors.Wrapf(err, "cannot create UDP connection on interface %s", inf.Name)
		return
	}
	if err = p.SetMulticastLoopback(true); err != nil {
		err = errors.Wrapf(err, "cannot set multicast loopback on interface %s", inf.Name)
		return
	}
	return
}

// SendUDP sends the message msg via connection conn to address addr
func SendUDP(conn *net.UDPConn, addr *net.UDPAddr, msg []byte) (err error) {
	var n int
	if n, err = conn.WriteToUDP(msg, addr); err != nil {
		err = errors.Wrap(err, "error writing to UDP socket")
		return
	} else if n != len(msg) {
		err = fmt.Errorf("incomplete write to UDP socket: %d/%d bytes", n, len(msg))
		return
	}

	return
}
