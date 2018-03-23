package vhost

import (
	"bufio"
	"net"
	"net/http"
)

type HTTPConn struct {
	*sharedConn
	Request *http.Request
}

// HTTP parses the head of the first HTTP request on conn and returns
// a new, unread connection with metadata for virtual host muxing
func HTTP(conn net.Conn) (httpConn *HTTPConn, err error) {
	c, rd := newShared(conn)

	httpConn = &HTTPConn{sharedConn: c}
	if httpConn.Request, err = http.ReadRequest(bufio.NewReader(rd)); err != nil {
		return
	}

	// You probably don't need access to the request body and this makes the API
	// simpler by allowing you to call Free() optionally
	httpConn.Request.Body.Close()

	return
}

// Free sets Request to nil so that it can be garbage collected
func (c *HTTPConn) Free() {
	c.Request = nil
}

func (c *HTTPConn) Host() string {
	if c.Request == nil {
		return ""
	}

	return c.Request.Host
}
