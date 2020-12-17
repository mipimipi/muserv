package events

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.com/mipimipi/yuppie/internal/network"
)

var (
	reURLs    = regexp.MustCompile(`^(<.+>)+$`)
	reTimeOut = regexp.MustCompile(`Second-\d+`)
)

// Subscription represents the subscription of one recipient to all evented
// state variables
type Subscription struct {
	sid       uuid.UUID
	timer     *time.Timer
	urls      []*url.URL
	stateVars []StateVar
	sequence  uint32
}

// sendEvent sends an event to the recipient of this subscription. As the UPnP
// Device Architecture 2.0 requires, it tries to send the message to all urls
// of that recipient subsequently as long as one these transmissions could be
// done without an error. Note: It's important that the very first event that
// is sent to a subscription has the the sequence id (SEQ) 0.
func (me *Subscription) sendEvent() {
	// assemble event message body
	body := marshalStatVars(me.stateVars)

	success := false
	for _, u := range me.urls {
		var err error

		// assemble message
		msg := new(bytes.Buffer)
		fmt.Fprintf(msg, "NOTIFY %s HTTP/1.1\r\n", u.Path)
		fmt.Fprintf(msg, "HOST: %s:%s\r\n", u.Hostname(), u.Port())
		fmt.Fprint(msg, "CONTENT-TYPE: text/xml; charset=\"utf-8\"\r\n")
		fmt.Fprintf(msg, "CONTENT-LENGTH: %d\r\n", len(body))
		fmt.Fprint(msg, "NT: upnp:event\r\n")
		fmt.Fprint(msg, "NTS: upnp:propchange\r\n")
		fmt.Fprintf(msg, "SID: uuid:%s\r\n", me.sid.String())
		fmt.Fprintf(msg, "SEQ: %d\r\n", me.sequence)
		fmt.Fprint(msg, "\r\n")
		fmt.Fprint(msg, string(body))
		fmt.Fprint(msg, "\r\n")

		// create TCP connection
		var conn net.Conn
		if conn, err = network.TCPConn(u.Hostname() + ":" + u.Port()); err != nil {
			log.Errorf("subscriptions: cannot create TCP connection to %s", u.Hostname()+":"+u.Port())
			continue
		}

		// send event message
		if err = network.SendTCP(conn, msg.Bytes()); err != nil {
			log.Infof("cannot send subscription event to %s", u.Hostname()+":"+u.Port())
			conn.Close()
			continue
		}
		conn.Close()

		log.Tracef("sent subscription event to %s, seq=%d ", u.String(), me.sequence)
		success = true
		break
	}
	if success {
		me.sequence++
	}
}

// ParseURLs parses the callback string that the recipient submitted as part of
// the subscription request. If the string is not according to the required
// format an error is returned. As defined in UPnP Device Architecture 2.0, the
// required format is <url_1><url_2>...<url_n>, where url_x must be a valid
// url for x=1, ..., n
func ParseURLs(callback string) (urls []*url.URL, err error) {
	// callback must be of the form <url_1><url_2>...<url_n>
	if !reURLs.MatchString(callback) {
		err = fmt.Errorf("callback malformatted: %s", callback)
		log.Error(err)
		return
	}

	// extract urls from callback
	callback = callback[1 : len(callback)-1]
	a := strings.Split(callback, "><")
	for _, s := range a {
		u, err := url.ParseRequestURI(s)
		if err != nil {
			err = fmt.Errorf("callback malformatted: %s", s)
			log.Error(err)
			return nil, err
		}
		urls = append(urls, u)
	}

	// at least one url required
	if len(urls) == 0 {
		err = fmt.Errorf("callback malformatted: %s", callback)
		log.Error(err)
		return nil, err
	}

	return
}

// ParseTimeout parses the timeout string that the recipient submitted as part
// of the subscription request. If the string is not according to the required
// format an error is returned. As defined in UPnP Device Architecture 2.0, the
// required format is Second-<number>, where <number> is requested timeout in
// seconds
func ParseTimeout(t string) (dur time.Duration, err error) {
	if t == "" {
		return
	}

	if !reTimeOut.MatchString(t) {
		err = errors.Wrapf(err, "timeout malformatted: %s", t)
		log.Error(err)
		return
	}

	f, err := strconv.ParseFloat(t[7:], 64)
	if err != nil {
		err = errors.Wrapf(err, "timeout malformatted: %s", t)
		log.Error(err)
		return
	}

	// make sure that minimum timeout is kept
	if time.Duration(f) < minSubTimeout*time.Second {
		dur = minSubTimeout * time.Second
	} else {
		dur = time.Duration(f)
	}

	return
}
