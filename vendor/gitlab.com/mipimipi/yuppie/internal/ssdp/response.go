package ssdp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/mipimipi/yuppie/internal/network"
)

// search targets
const (
	stAll  = "ssdp:all"
	stRoot = "upnp:rootdevice"
)

// SearchIndex maps keys like service and device types to the USNs that must be
// sent as response to search requests
type SearchIndex map[string](*([]string))

var reTypes = regexp.MustCompile(`urn:.+:(device|service):.+:.+`)

// isCompatible checks if a and b are compatible version strings. I.e. they
// represent the same device of service type and the version of a is less or
// equal to the version of b
func isCompatible(a, b string) bool {
	if !reTypes.MatchString(a) || !reTypes.MatchString(b) {
		return false
	}
	n := strings.LastIndex(a, ":")
	return n+1 < len(a) && n+1 < len(b) && a[:n] == b[:n] && a[n+1:] <= b[n+1:]
}

// retrieve searches in me for a key that either matches target or if there is
// a key that is a device or service type which is compatible to target.
func (me SearchIndex) retrieve(target string) (*[]string, bool) {
	for key := range me {
		if key == target {
			return me[key], true
		}
		if reTypes.MatchString(target) {
			if isCompatible(target, key) {
				return me[key], true
			}
		}
	}

	return nil, false
}

// listenAndRespond listens to SSDP search requests and responds to them if
// they are relevant
func (me *Server) listenAndRespond() {
	me.responseStopped = make(chan struct{})

	// loop for receiving search requests
	var wg sync.WaitGroup
	for {
		select {
		case <-me.stopResponse:
			// wait until a response that might just be sent is out
			wg.Wait()
			close(me.responseStopped)
			log.Tracef("response stopped on interface '%s'", me.inf.Name)
			return
		default:
			// read data from UDP connection
			msg := make([]byte, me.inf.MTU)
			n, reqAddr, err := me.conn.ReadFromUDP(msg)
			if err != nil {
				err = fmt.Errorf("search: error reading from UDP socket: %v", err)
				log.Error(err)
				continue
			}
			if n > 0 {
				// request received: respond in go routine to not block reading of
				// further messages
				wg.Add(1)
				go me.respond(&wg, msg[:n], reqAddr)
			}
		}
	}
}

// respond evaluates a search request (msg) and responds if the request is
// relevant
func (me *Server) respond(wg *sync.WaitGroup, msg []byte, reqAddr *net.UDPAddr) {
	defer wg.Done()

	// transform msg into HTTP request struct
	r, err := parseIntoHTTPRequest(bufio.NewReader(bytes.NewReader(msg)))
	if err != nil {
		// msg is either not a search request or it is mal-formed. In both
		// cases the UPnP Device Architecture 2.0 spec required to silently
		// ignore the request
		log.Infof("cannot create HTTP request from search request: %v", err)
		return
	}

	// analyze msg and extract data for search response
	st, mx, tcpPort, isRelevant, err := analyzeHTTPRequest(r, me.index)
	if err != nil {
		// msg is either not a search request or it is mal-formed. In both
		// cases the UPnP Device Architecture 2.0 spec required to silently
		// ignore the request
		//log.Infof("cannot analyze HTTP request: %v", err)
		return
	}

	// here we know that msg is a well-formed search request: respond if search
	// is relevant
	if isRelevant {
		log.Tracef("search request from %s for %s on interface %s is relevant", reqAddr.IP.String(), st, me.inf.Name)
		_ = me.sendResponse(st, mx, reqAddr, tcpPort)
	}
}

// sendResponse sends a response for a search request
func (me *Server) sendResponse(st string, mx uint, reqAddr *net.UDPAddr, tcpPort int) (err error) {
	// assemble response messages
	msgs := me.assembleResponseMsgs(st, (tcpPort != 0))

	// as the UPnP Device Architecture 2.0 spec says: A tcpPort != 0 means to
	// send the responses per TCP, otherwise per UDP
	if tcpPort == 0 {
		// send messages via UDP spread over a time intervall of mx seconds (as
		// required by UPnP Device Architecture 2.0)
		for i := 0; i < len(msgs); i++ {
			if i != 0 {
				time.Sleep(time.Duration(mx) * time.Second / time.Duration(len(msgs)+1))
			}
			if err = network.SendUDP(me.conn, reqAddr, msgs[i].Bytes()); err != nil {
				err = errors.Wrap(err, "couldn't send SSDP search response")
				log.Error(err)
			}
		}
	} else {
		// send messages via TCP. As per the UPnP Device Architecture 2.0
		// specification, all messages can be sent at once

		// assemble target address
		ip := reqAddr.IP.To4()
		if ip == nil {
			err = fmt.Errorf("search response: addr '%s' is not a IPv4 address", reqAddr.IP.String())
			log.Error(err)
			return
		}

		// create TCP connection
		var conn net.Conn
		if conn, err = network.TCPConn(ip.String() + ":" + strconv.Itoa(tcpPort)); err != nil {
			err = fmt.Errorf("search response: cannot create TCP connection to %s", ip.String()+":"+strconv.Itoa(tcpPort))
			log.Error(err)
			return
		}
		defer conn.Close()

		// send messages
		for _, msg := range msgs {
			_ = network.SendTCP(conn, msg.Bytes())
		}
	}

	log.Infof("responded to search request from %s for %s on interface %s", reqAddr.IP.String(), st, me.inf.Name)
	return
}

// assembleResponseMsgs create the message texts for a response to a search request
func (me *Server) assembleResponseMsgs(st string, tcpRequired bool) (msgs []*bytes.Buffer) {
	switch st {
	case stAll:
		for _, assID := range me.data.AssIDs {
			msg := new(bytes.Buffer)
			fmt.Fprint(msg, "HTTP/1.1 200 OK\r\n")
			fmt.Fprintf(msg, "CACHE-CONTROL: max-age=%d\r\n", me.data.MaxAge)
			fmt.Fprintf(msg, "DATE: %s\r\n", time.Now().Format(time.RFC1123))
			fmt.Fprintf(msg, "EXT:\r\n")
			fmt.Fprintf(msg, "LOCATION: %s\r\n", strings.Replace(me.data.Location, "{{ADDRESS}}", me.addr, -1))
			fmt.Fprintf(msg, "SERVER: %s\r\n", me.data.Server)
			fmt.Fprintf(msg, "ST: %s\r\n", stAll)
			fmt.Fprintf(msg, "USN: %s\r\n", assID.USN)
			fmt.Fprintf(msg, "BOOTID.UPNP.ORG: %d\r\n", me.bootID.Val())
			fmt.Fprintf(msg, "CONFIG.UPNP.ORG: %d\r\n", me.configID.Val())
			// add empty row at the end as required by the UPnP Device
			// Architecture 2.0
			fmt.Fprint(msg, "\r\n")
			msgs = append(msgs, msg)
		}

	case stRoot:
		// get usn for root device
		usns, ok := me.index[stRoot]
		if !ok {
			log.Error("search response: root device not found in search index")
			return
		}
		if len(*usns) != 1 {
			log.Errorf("search response: for key '%s' more than one device is contained in search index", stRoot)
			return
		}
		msg := new(bytes.Buffer)
		fmt.Fprint(msg, "HTTP/1.1 200 OK\r\n")
		fmt.Fprintf(msg, "CACHE-CONTROL: max-age=%d\r\n", me.data.MaxAge)
		fmt.Fprintf(msg, "DATE: %s\r\n", time.Now().Format(time.RFC1123))
		fmt.Fprintf(msg, "EXT:\r\n")
		fmt.Fprintf(msg, "LOCATION: %s\r\n", strings.Replace(me.data.Location, "{{ADDRESS}}", me.addr, -1))
		fmt.Fprintf(msg, "SERVER: %s\r\n", me.data.Server)
		fmt.Fprintf(msg, "ST: %s\r\n", stRoot)
		fmt.Fprintf(msg, "USN: %s\r\n", (*usns)[0])
		fmt.Fprintf(msg, "BOOTID.UPNP.ORG: %d\r\n", me.bootID.Val())
		fmt.Fprintf(msg, "CONFIG.UPNP.ORG: %d\r\n", me.configID.Val())
		// add empty row at the end as required by the UPnP Device
		// Architecture 2.0
		fmt.Fprint(msg, "\r\n")
		msgs = append(msgs, msg)

	default:
		// get usns that must be sent
		usns, ok := me.index.retrieve(st)
		if !ok || len(*usns) == 0 {
			// apparently, search request was for a specific device, device
			// type or service type, but did not fit to the device tree:
			// nothing to do
			return
		}

		// assemble messages for UDP or TCP
		if !tcpRequired {
			// as specified in the UPnP Device Architecture 2.0, the search
			// response must be sent via UDP - one message per USN. If the
			// message shall be sent via TCP and there's only one USN, the same
			// logic applies
			for _, usn := range *usns {
				msg := new(bytes.Buffer)
				fmt.Fprint(msg, "HTTP/1.1 200 OK\r\n")
				fmt.Fprintf(msg, "CACHE-CONTROL: max-age=%d\r\n", me.data.MaxAge)
				fmt.Fprintf(msg, "DATE: %s\r\n", time.Now().Format(time.RFC1123))
				fmt.Fprintf(msg, "EXT:\r\n")
				fmt.Fprintf(msg, "LOCATION: %s\r\n", strings.Replace(me.data.Location, "{{ADDRESS}}", me.addr, -1))
				fmt.Fprintf(msg, "SERVER: %s\r\n", me.data.Server)
				fmt.Fprintf(msg, "ST: %s\r\n", st)
				fmt.Fprintf(msg, "USN: %s\r\n", usn)
				fmt.Fprintf(msg, "BOOTID.UPNP.ORG: %d\r\n", me.bootID.Val())
				fmt.Fprintf(msg, "CONFIG.UPNP.ORG: %d\r\n", me.configID.Val())
				// add empty row at the end as required by the UPnP Device
				// Architecture 2.0
				fmt.Fprint(msg, "\r\n")

				msgs = append(msgs, msg)
			}
		} else {
			// as specified in the UPnP Device Architecture 2.0, the search
			// response must be sent via TCP - one message in total, the
			// different USNs (if there are more than one) are sent as
			// comma-separated list in the USN field
			msg := new(bytes.Buffer)
			fmt.Fprint(msg, "HTTP/1.1 200 OK\r\n")
			fmt.Fprintf(msg, "CACHE-CONTROL: max-age=%d\r\n", me.data.MaxAge)
			fmt.Fprintf(msg, "DATE: %s\r\n", time.Now().Format(time.RFC1123))
			fmt.Fprintf(msg, "EXT:\r\n")
			fmt.Fprintf(msg, "LOCATION: %s\r\n", strings.Replace(me.data.Location, "{{ADDRESS}}", me.addr, -1))
			fmt.Fprintf(msg, "SERVER: %s\r\n", me.data.Server)
			fmt.Fprintf(msg, "ST: %s\r\n", st)
			for i := 0; i < len(*usns); i++ {
				if i == 0 {
					fmt.Fprintf(msg, "USN: %s", (*usns)[i])
					continue
				}
				fmt.Fprintf(msg, ",%s", (*usns)[i])
			}
			fmt.Fprintf(msg, "\r\n")
			fmt.Fprintf(msg, "BOOTID.UPNP.ORG: %d\r\n", me.bootID.Val())
			fmt.Fprintf(msg, "CONFIG.UPNP.ORG: %d\r\n", me.configID.Val())
			// add empty row at the end as required by the UPnP Device
			// Architecture 2.0
			fmt.Fprint(msg, "\r\n")

			msgs = append(msgs, msg)
		}
	}

	return
}

// analyzeHTTPRequest evaluates a search request and checks if it is relevant.
// isRelevant is set accordingly. If the request is relevant, st, mx and
// tcpPort are filled with the corresponding request values
func analyzeHTTPRequest(r *http.Request, index SearchIndex) (st string, mx uint, tcpPort int, isRelevant bool, err error) {
	// analyze request data
	// - method
	if r.Method != "M-SEARCH" {
		err = fmt.Errorf("search request: wrong method: %s", r.Method)
		return
	}
	// - MAN field
	if r.Header.Get("MAN") != `"ssdp:discover"` {
		err = fmt.Errorf("search request: wrong MAN field: %s", r.Header.Get("MAN"))
		return
	}

	// - multicast request?
	var isMulticast bool
	if r.Header.Get("HOST") == multicastAddrIPv4 {
		isMulticast = true
	}

	// - MX field
	if isMulticast {
		// MX is only required for a multicast search request
		mxHeader := r.Header.Get("MX")
		var u uint64
		u, err = strconv.ParseUint(mxHeader, 0, 0)
		if err != nil {
			err = errors.Wrapf(err, "search: invalid MX header %q: %v", mxHeader, err)
			return
		}
		// according to the UPnP Device Architecture 2.0 spec, MX shall be
		// 5 seconds max
		if u > 5 {
			mx = 5
		} else {
			mx = uint(u) // note: ParseUint returns uint64 and not uint
		}
	} else {
		// unicast search request: set MX to default value
		mx = 1
	}
	// - ST field: check is request is relevant at all
	st = r.Header.Get("ST")
	isRelevant = (st == stAll || st == stRoot)
	if !isRelevant {
		_, isRelevant = index.retrieve(r.Header.Get("ST"))
	}
	if !isRelevant {
		return
	}
	// - TCP port
	if r.Header.Get("TCPPORT.UPNP.ORG") != "" {
		if tcpPort, err = strconv.Atoi(r.Header.Get("TCPPORT.UPNP.ORG")); err != nil {
			tcpPort = 0
		}
	}

	return
}

// parseIntoHTTPRequest takes the message text of a search request and and
// creates a HTTP request from it for further analysis
func parseIntoHTTPRequest(msg *bufio.Reader) (r *http.Request, err error) {
	tp := textproto.NewReader(msg)
	var s string
	if s, err = tp.ReadLine(); err != nil {
		err = errors.Wrap(err, "cannot read from message")
		return
	}
	defer func() {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()

	// analyze first request line
	var line []string
	if line = strings.SplitN(s, " ", 3); len(line) < 3 {
		err = fmt.Errorf("search: malformed request line: %s", s)
		return nil, err
	}
	if line[1] != "*" {
		err = fmt.Errorf("search: bad URL request: %s", line[1])
		return nil, err
	}

	// assemble HTTP request struct
	r = &http.Request{Method: line[0]}
	var ok bool
	if r.ProtoMajor, r.ProtoMinor, ok = http.ParseHTTPVersion(strings.TrimSpace(line[2])); !ok {
		err = errors.Wrapf(err, "search: malformed HTTP version: %s", line[2])
		return nil, err
	}
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	r.Header = http.Header(mimeHeader)

	return
}
