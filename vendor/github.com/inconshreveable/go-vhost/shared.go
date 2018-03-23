package vhost

import (
	"bytes"
	"io"
	"net"
	"sync"
)

const (
	initVhostBufSize = 1024 // allocate 1 KB up front to try to avoid resizing
)

type sharedConn struct {
	sync.Mutex
	net.Conn               // the raw connection
	vhostBuf *bytes.Buffer // all of the initial data that has to be read in order to vhost a connection is saved here
}

func newShared(conn net.Conn) (*sharedConn, io.Reader) {
	c := &sharedConn{
		Conn:     conn,
		vhostBuf: bytes.NewBuffer(make([]byte, 0, initVhostBufSize)),
	}

	return c, io.TeeReader(conn, c.vhostBuf)
}

func (c *sharedConn) Read(p []byte) (n int, err error) {
	c.Lock()
	if c.vhostBuf == nil {
		c.Unlock()
		return c.Conn.Read(p)
	}
	n, err = c.vhostBuf.Read(p)

	// end of the request buffer
	if err == io.EOF {
		// let the request buffer get garbage collected
		// and make sure we don't read from it again
		c.vhostBuf = nil

		// continue reading from the connection
		var n2 int
		n2, err = c.Conn.Read(p[n:])

		// update total read
		n += n2
	}
	c.Unlock()
	return
}
