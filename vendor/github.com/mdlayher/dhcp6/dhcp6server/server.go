package dhcp6server

import (
	"errors"
	"log"
	"net"

	"github.com/mdlayher/dhcp6"
	"github.com/mdlayher/dhcp6/dhcp6opts"
	"golang.org/x/net/ipv6"
)

var (
	// AllRelayAgentsAndServersAddr is the multicast address group which is
	// used to communicate with neighboring (on-link) DHCP servers and relay
	// agents, as defined in RFC 3315, Section 5.1.  All DHCP servers
	// and relay agents are members of this multicast group.
	AllRelayAgentsAndServersAddr = &net.IPAddr{
		IP: net.ParseIP("ff02::1:2"),
	}

	// AllServersAddr is the multicast address group which is used by a
	// DHCP relay agent to communicate with DHCP servers, if the relay agent
	// wishes to send messages to all servers, or does not know the unicast
	// address of a server.  All DHCP servers are members of this multicast
	// group.
	AllServersAddr = &net.IPAddr{
		IP: net.ParseIP("ff05::1:3"),
	}

	// errClosing is a special value used to stop the server's read loop
	// when a connection is closing.
	errClosing = errors.New("use of closed network connection")
)

// PacketConn is an interface which types must implement in order to serve
// DHCP connections using Server.Serve.
type PacketConn interface {
	ReadFrom(b []byte) (n int, cm *ipv6.ControlMessage, src net.Addr, err error)
	WriteTo(b []byte, cm *ipv6.ControlMessage, dst net.Addr) (n int, err error)

	Close() error

	JoinGroup(ifi *net.Interface, group net.Addr) error
	LeaveGroup(ifi *net.Interface, group net.Addr) error

	SetControlMessage(cf ipv6.ControlFlags, on bool) error
}

// Server represents a DHCP server, and is used to configure a DHCP server's
// behavior.
type Server struct {
	// Iface is the the network interface on which this server should
	// listen.  Traffic from any other network interface will be filtered out
	// and ignored by the server.
	Iface *net.Interface

	// Addr is the network address which this server should bind to.  The
	// default value is [::]:547, as specified in RFC 3315, Section 5.2.
	Addr string

	// Handler is the handler to use while serving DHCP requests.  If this
	// value is nil, the Server will panic.
	Handler Handler

	// MulticastGroups designates which IPv6 multicast groups this server
	// will join on start-up.  Because the default configuration acts as a
	// DHCP server, most servers will typically join both
	// AllRelayAgentsAndServersAddr, and AllServersAddr. If configuring a
	// DHCP relay agent, only the former value should be used.
	MulticastGroups []*net.IPAddr

	// ServerID is the the server's DUID, which uniquely identifies this
	// server to clients.  If no DUID is specified, a DUID-LL will be
	// generated using Iface's hardware type and address.  If possible,
	// servers with persistent storage available should generate a DUID-LLT
	// and store it for future use.
	ServerID dhcp6opts.DUID

	// ErrorLog is an optional logger which can be used to report errors and
	// erroneous behavior while the server is accepting client requests.
	// If ErrorLog is nil, logging goes to os.Stderr via the log package's
	// standard logger.
	ErrorLog *log.Logger
}

// logf logs a message using the server's ErrorLog logger, or the log package
// standard logger, if ErrorLog is nil.
func (s *Server) logf(format string, args ...interface{}) {
	if s.ErrorLog != nil {
		s.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// ListenAndServe listens for UDP6 connections on the specified address of the
// specified interface, using the default Server configuration and specified
// handler to handle DHCPv6 connections.  The Handler must not be nil.
//
// Any traffic which reaches the Server, and is not bound for the specified
// network interface, will be filtered out and ignored.
//
// In this configuration, the server acts as a DHCP server, but NOT as a
// DHCP relay agent.  For more information on DHCP relay agents, see RFC 3315,
// Section 20.
func ListenAndServe(iface string, handler Handler) error {
	// Verify network interface exists
	ifi, err := net.InterfaceByName(iface)
	if err != nil {
		return err
	}

	return (&Server{
		Iface:   ifi,
		Addr:    "[::]:547",
		Handler: handler,
		MulticastGroups: []*net.IPAddr{
			AllRelayAgentsAndServersAddr,
			AllServersAddr,
		},
	}).ListenAndServe()
}

// ListenAndServe listens on the address specified by s.Addr using the network
// interface defined in s.Iface.  Traffic from any other interface will be
// filtered out and ignored.  Serve is called to handle serving DHCP traffic
// once ListenAndServe opens a UDP6 packet connection.
func (s *Server) ListenAndServe() error {
	// Open UDP6 packet connection listener on specified address
	conn, err := net.ListenPacket("udp6", s.Addr)
	if err != nil {
		return err
	}

	defer conn.Close()
	return s.Serve(ipv6.NewPacketConn(conn))
}

// Serve configures and accepts incoming connections on PacketConn p, creating a
// new goroutine for each.  Serve configures IPv6 control message settings, joins
// the appropriate multicast groups, and begins listening for incoming connections.
//
// The service goroutine reads requests, generate the appropriate Request and
// ResponseSender values, then calls s.Handler to handle the request.
func (s *Server) Serve(p PacketConn) error {
	// If no DUID was set for server previously, generate a DUID-LL
	// now using the interface's hardware address, and just assume the
	// "Ethernet 10Mb" hardware type since the caller probably doesn't care.
	if s.ServerID == nil {
		const ethernet10Mb uint16 = 1
		s.ServerID = dhcp6opts.NewDUIDLL(ethernet10Mb, s.Iface.HardwareAddr)
	}

	// Filter any traffic which does not indicate the interface
	// defined by s.Iface.
	if err := p.SetControlMessage(ipv6.FlagInterface, true); err != nil {
		return err
	}

	// Join appropriate multicast groups
	for _, g := range s.MulticastGroups {
		if err := p.JoinGroup(s.Iface, g); err != nil {
			return err
		}
	}

	// Set up IPv6 packet connection, and on return, handle leaving multicast
	// groups and closing connection
	defer func() {
		for _, g := range s.MulticastGroups {
			_ = p.LeaveGroup(s.Iface, g)
		}

		_ = p.Close()
	}()

	// Loop and read requests until exit
	buf := make([]byte, 1500)
	for {
		n, cm, addr, err := p.ReadFrom(buf)
		if err != nil {
			// Stop serve loop gracefully when closing
			if err == errClosing {
				return nil
			}

			// BUG(mdlayher): determine if error can be temporary
			return err
		}

		// Filter any traffic with a control message indicating an incorrect
		// interface index
		if cm != nil && cm.IfIndex != s.Iface.Index {
			continue
		}

		// Create conn struct with data specific to this connection
		uc, err := s.newConn(p, addr.(*net.UDPAddr), n, buf)
		if err != nil {
			continue
		}

		// Serve conn and continue looping for more connections
		go uc.serve()
	}
}

// conn represents an in-flight DHCP connection, and contains information about
// the connection and server.
type conn struct {
	conn       PacketConn
	remoteAddr *net.UDPAddr
	server     *Server
	buf        []byte
}

// newConn creates a new conn using information received in a single DHCP
// connection.  newConn makes a copy of the input buffer for use in handling
// a single connection.
// BUG(mdlayher): consider using a sync.Pool with many buffers available to avoid
// allocating a new one on each connection
func (s *Server) newConn(p PacketConn, addr *net.UDPAddr, n int, buf []byte) (*conn, error) {
	c := &conn{
		conn:       p,
		remoteAddr: addr,
		server:     s,
		buf:        make([]byte, n, n),
	}
	copy(c.buf, buf[:n])

	return c, nil
}

// response represents a DHCP response, and implements ResponseSender so that
// outbound Packets can be appropriately created and sent to a client.
type response struct {
	conn       PacketConn
	remoteAddr *net.UDPAddr
	req        *Request

	options dhcp6.Options
}

// Options returns the Options map, which can be modified before a call
// to Write.  When Write is called, the Options map is enumerated into an
// ordered slice of option codes and values.
func (r *response) Options() dhcp6.Options {
	return r.options
}

// Send uses the input message typ, the transaction ID sent by a client,
// and the options set by Options, to create and send a Packet to the
// client's address.
func (r *response) Send(mt dhcp6.MessageType) (int, error) {
	p := &dhcp6.Packet{
		MessageType:   mt,
		TransactionID: r.req.TransactionID,
		Options:       r.options,
	}

	b, err := p.MarshalBinary()
	if err != nil {
		return 0, err
	}

	return r.conn.WriteTo(b, nil, r.remoteAddr)
}

// serve handles serving an individual DHCP connection, and is invoked in a
// goroutine.
func (c *conn) serve() {
	// Attempt to parse a Request from a raw packet, providing a nicer
	// API for callers to implement their own DHCP request handlers.
	r, err := ParseRequest(c.buf, c.remoteAddr)
	if err != nil {
		// Malformed packets get no response
		if err == dhcp6.ErrInvalidPacket {
			return
		}

		// BUG(mdlayher): decide to log or handle other request errors
		c.server.logf("%s: error parsing request: %s", c.remoteAddr.String(), err.Error())
		return
	}

	// Filter out unknown/invalid message types, using the lowest and highest
	// numbered types
	if r.MessageType < dhcp6.MessageTypeSolicit || r.MessageType > dhcp6.MessageTypeDHCPv4Response {
		c.server.logf("%s: unrecognized message type: %d", c.remoteAddr.String(), r.MessageType)
		return
	}

	// Set up response to send responses back to the original requester
	w := &response{
		remoteAddr: c.remoteAddr,
		conn:       c.conn,
		req:        r,
		options:    make(dhcp6.Options),
	}

	// Add server ID to response
	if sID := c.server.ServerID; sID != nil {
		_ = w.options.Add(dhcp6.OptionServerID, sID)
	}

	// If available in request, add client ID to response
	if cID, err := dhcp6opts.GetClientID(r.Options); err == nil {
		w.options.Add(dhcp6.OptionClientID, cID)
	}

	// Enforce a valid Handler.
	handler := c.server.Handler
	if handler == nil {
		panic("nil DHCPv6 handler for server")
	}

	handler.ServeDHCP(w, r)
}
