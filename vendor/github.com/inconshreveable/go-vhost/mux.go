package vhost

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	normalize = strings.ToLower
	isClosed  = func(err error) bool {
		netErr, ok := err.(net.Error)
		if ok {
			return netErr.Temporary()
		}
		return false
	}
)

// NotFound is returned when a vhost is not found
type NotFound struct {
	error
}

// BadRequest is returned when extraction of the vhost name fails
type BadRequest struct {
	error
}

// Closed is returned when the underlying connection is closed
type Closed struct {
	error
}

type (
	// this is the function you apply to a net.Conn to get
	// a new virtual-host multiplexed connection
	muxFn func(net.Conn) (Conn, error)

	// an error encountered when multiplexing a connection
	muxErr struct {
		err  error
		conn net.Conn
	}
)

type VhostMuxer struct {
	listener     net.Listener         // listener on which we mux connections
	muxTimeout   time.Duration        // a connection fails if it doesn't send enough data to mux after this timeout
	vhostFn      muxFn                // new connections are multiplexed by applying this function
	muxErrors    chan muxErr          // all muxing errors are sent over this channel
	registry     map[string]*Listener // registry of name -> listener
	sync.RWMutex                      // protects the registry
}

func NewVhostMuxer(listener net.Listener, vhostFn muxFn, muxTimeout time.Duration) (*VhostMuxer, error) {
	mux := &VhostMuxer{
		listener:   listener,
		muxTimeout: muxTimeout,
		vhostFn:    vhostFn,
		muxErrors:  make(chan muxErr),
		registry:   make(map[string]*Listener),
	}

	go mux.run()
	return mux, nil
}

// Listen begins multiplexing the underlying connection to send new
// connections for the given name over the returned listener.
func (m *VhostMuxer) Listen(name string) (net.Listener, error) {
	name = normalize(name)

	vhost := &Listener{
		name:   name,
		mux:    m,
		accept: make(chan Conn),
	}

	if err := m.set(name, vhost); err != nil {
		return nil, err
	}

	return vhost, nil
}

// NextError returns the next error encountered while mux'ing a connection.
// The net.Conn may be nil if the wrapped listener returned an error from Accept()
func (m *VhostMuxer) NextError() (net.Conn, error) {
	muxErr := <-m.muxErrors
	return muxErr.conn, muxErr.err
}

// Close closes the underlying listener
func (m *VhostMuxer) Close() {
	m.listener.Close()
}

// run is the VhostMuxer's main loop for accepting new connections from the wrapped listener
func (m *VhostMuxer) run() {
	for {
		conn, err := m.listener.Accept()
		if err != nil {
			if isClosed(err) {
				m.sendError(nil, Closed{err})
				return
			} else {
				m.sendError(nil, err)
				continue
			}
		}
		go m.handle(conn)
	}
}

// handle muxes a connection accepted from the listener
func (m *VhostMuxer) handle(conn net.Conn) {
	defer func() {
		// recover from failures
		if r := recover(); r != nil {
			m.sendError(conn, fmt.Errorf("NameMux.handle failed with error %v", r))
		}
	}()

	// Make sure we detect dead connections while we decide how to multiplex
	if err := conn.SetDeadline(time.Now().Add(m.muxTimeout)); err != nil {
		m.sendError(conn, fmt.Errorf("Failed to set deadline: %v", err))
		return
	}

	// extract the name
	vconn, err := m.vhostFn(conn)
	if err != nil {
		m.sendError(conn, BadRequest{fmt.Errorf("Failed to extract vhost name: %v", err)})
		return
	}

	// normalize the name
	host := normalize(vconn.Host())

	// look up the correct listener
	l, ok := m.get(host)
	if !ok {
		m.sendError(vconn, NotFound{fmt.Errorf("Host not found: %v", host)})
		return
	}

	if err = vconn.SetDeadline(time.Time{}); err != nil {
		m.sendError(vconn, fmt.Errorf("Failed unset connection deadline: %v", err))
		return
	}

	l.accept <- vconn
}

func (m *VhostMuxer) sendError(conn net.Conn, err error) {
	m.muxErrors <- muxErr{conn: conn, err: err}
}

func (m *VhostMuxer) get(name string) (l *Listener, ok bool) {
	m.RLock()
	defer m.RUnlock()
	l, ok = m.registry[name]
	if !ok {
		// look for a matching wildcard
		parts := strings.Split(name, ".")
		for i := 0; i < len(parts)-1; i++ {
			parts[i] = "*"
			name = strings.Join(parts[i:], ".")
			l, ok = m.registry[name]
			if ok {
				break
			}
		}
	}
	return
}

func (m *VhostMuxer) set(name string, l *Listener) error {
	m.Lock()
	defer m.Unlock()
	if _, exists := m.registry[name]; exists {
		return fmt.Errorf("name %s is already bound", name)
	}
	m.registry[name] = l
	return nil
}

func (m *VhostMuxer) del(name string) {
	m.Lock()
	defer m.Unlock()
	delete(m.registry, name)
}

const (
	serverError = `HTTP/1.0 500 Internal Server Error
Content-Length: 22

Internal Server Error
`

	notFound = `HTTP/1.0 404 Not Found
Content-Length: 14

404 not found
`

	badRequest = `HTTP/1.0 400 Bad Request
Content-Length: 12

Bad Request
`
)

type HTTPMuxer struct {
	*VhostMuxer
}

// HandleErrors handles muxing errors by calling .NextError(). You must
// invoke this function if you do not want to handle the errors yourself.
func (m *HTTPMuxer) HandleErrors() {
	for {
		m.HandleError(m.NextError())
	}
}

func (m *HTTPMuxer) HandleError(conn net.Conn, err error) {
	switch err.(type) {
	case Closed:
		return
	case NotFound:
		conn.Write([]byte(notFound))
	case BadRequest:
		conn.Write([]byte(badRequest))
	default:
		if conn != nil {
			conn.Write([]byte(serverError))
		}
	}

	if conn != nil {
		conn.Close()
	}
}

// NewHTTPMuxer begins muxing HTTP connections on the given listener by inspecting
// the HTTP Host header in new connections.
func NewHTTPMuxer(listener net.Listener, muxTimeout time.Duration) (*HTTPMuxer, error) {
	fn := func(c net.Conn) (Conn, error) { return HTTP(c) }
	mux, err := NewVhostMuxer(listener, fn, muxTimeout)
	return &HTTPMuxer{mux}, err
}

type TLSMuxer struct {
	*VhostMuxer
}

// HandleErrors is the default error handler for TLS muxers. At the moment, it simply
// closes connections which are invalid or destined for virtual host names that it is
// not listening for.
// You must invoke this function if you do not want to handle the errors yourself.
func (m *TLSMuxer) HandleErrors() {
	for {
		conn, err := m.NextError()

		if conn == nil {
			if _, ok := err.(Closed); ok {
				return
			} else {
				continue
			}
		} else {
			// XXX: respond with valid TLS close messages
			conn.Close()
		}
	}
}

func (m *TLSMuxer) Listen(name string) (net.Listener, error) {
	// TLS SNI never includes the port
	host, _, err := net.SplitHostPort(name)
	if err != nil {
		host = name
	}
	return m.VhostMuxer.Listen(host)
}

// NewTLSMuxer begins muxing TLS connections by inspecting the SNI extension.
func NewTLSMuxer(listener net.Listener, muxTimeout time.Duration) (*TLSMuxer, error) {
	fn := func(c net.Conn) (Conn, error) { return TLS(c) }
	mux, err := NewVhostMuxer(listener, fn, muxTimeout)
	return &TLSMuxer{mux}, err
}

// Listener is returned by a call to Listen() on a muxer. A Listener
// only receives connections that were made to the name passed into the muxer's
// Listen call.
//
// Listener implements the net.Listener interface, so you can Accept() new
// connections and Close() it when finished. When you Close() a Listener,
// the parent muxer will stop listening for connections to the Listener's name.
type Listener struct {
	name   string
	mux    *VhostMuxer
	accept chan Conn
}

// Accept returns the next mux'd connection for this listener and blocks
// until one is available.
func (l *Listener) Accept() (net.Conn, error) {
	conn, ok := <-l.accept
	if !ok {
		return nil, fmt.Errorf("Listener closed")
	}
	return conn, nil
}

// Close stops the parent muxer from listening for connections to the mux'd
// virtual host name.
func (l *Listener) Close() error {
	l.mux.del(l.name)
	close(l.accept)
	return nil
}

// Addr returns the address of the bound listener used by the parent muxer.
func (l *Listener) Addr() net.Addr {
	// XXX: include name in address?
	return l.mux.listener.Addr()
}

// Name returns the name of the virtual host this listener receives connections on.
func (l *Listener) Name() string {
	return l.name
}
