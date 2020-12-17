package network

import (
	"fmt"
	"net"

	"github.com/pkg/errors"
)

// TCPConn creates a TCP network connection to address addr
func TCPConn(addr string) (conn net.Conn, err error) {
	conn, err = net.Dial("tcp4", addr)
	if err != nil {
		err = errors.Wrapf(err, "cannot create TCP connection to address %s", addr)
		return
	}
	return
}

// SendTCP sends the message msg via connection conn
func SendTCP(conn net.Conn, msg []byte) (err error) {
	var n int
	n, err = conn.Write(msg)
	if err != nil {
		err = errors.Wrap(err, "TCP message could not be sent")
		return
	}
	if n < len(msg) {
		err = fmt.Errorf("TCP message could not be sent completely: %d/%d bytes", n, len(msg))
		return
	}
	return
}
