package dhcp6server

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/mdlayher/dhcp6"
	"github.com/mdlayher/dhcp6/dhcp6opts"
	"golang.org/x/net/ipv6"
)

func init() {
	// Discard all output from Server.logf
	log.SetOutput(ioutil.Discard)
}

// TestServeIPv6ControlParameters verifies that a PacketConn successfully
// joins and leaves the appropriate multicast groups designated by a Server,
// and that the appropriate IPv6 control flags are set.
func TestServeIPv6ControlParameters(t *testing.T) {
	s := &Server{
		MulticastGroups: []*net.IPAddr{
			AllRelayAgentsAndServersAddr,
			AllServersAddr,
		},
	}

	// Send pseudo-Packet to avoid EOF, even though it does not matter
	// for this test
	r := &testMessage{}
	r.b.Write([]byte{0, 0, 0, 0})

	// Don't expect a reply, don't handle a request
	_, ip6, err := testServe(r, s, false, func(w ResponseSender, r *Request) {})
	if err != nil {
		t.Fatal(err)
	}

	for _, m := range s.MulticastGroups {
		var foundJ bool
		var foundL bool

		for _, j := range ip6.joined {
			if m.String() == j.String() {
				foundJ = true
				break
			}
		}
		for _, l := range ip6.left {
			if m.String() == l.String() {
				foundL = true
				break
			}
		}

		if !foundJ {
			t.Fatalf("did not find joined IPv6 multicast group: %v", m)
		}
		if !foundL {
			t.Fatalf("did not find left IPv6 multicast group: %v", m)
		}
	}

	if b, ok := ip6.flags[ipv6.FlagInterface]; !ok || !b {
		t.Fatalf("FlagInterface not found or not set to true:\n- found: %v\n-  bool: %v", ok, b)
	}

	if !ip6.closed {
		t.Fatal("IPv6 connection not closed after Serve")
	}
}

// TestServeWithSetServerID verifies that Serve uses the server ID provided
// instead of generating its own, when a server ID is set.
func TestServeWithSetServerID(t *testing.T) {
	p := &dhcp6.Packet{
		MessageType:   dhcp6.MessageTypeSolicit,
		TransactionID: [3]byte{0, 1, 2},
	}

	pb, err := p.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	r := &testMessage{}
	r.b.Write(pb)

	duid, err := dhcp6opts.NewDUIDLLT(1, time.Now(), []byte{0, 1, 0, 1, 0, 1})
	if err != nil {
		t.Fatal(err)
	}

	s := &Server{
		ServerID: duid,
	}

	// Expect a reply with type advertise
	mt := dhcp6.MessageTypeAdvertise
	w, _, err := testServe(r, s, true, func(w ResponseSender, r *Request) {
		w.Send(mt)
	})
	if err != nil {
		t.Fatal(err)
	}

	wp := new(dhcp6.Packet)
	if err := wp.UnmarshalBinary(w.b.Bytes()); err != nil {
		t.Fatal(err)
	}

	if want, got := mt, wp.MessageType; want != got {
		t.Fatalf("unexpected message type: %v != %v", want, got)
	}

	want, err := duid.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	got, err := wp.Options.GetOne(dhcp6.OptionServerID)
	if err != nil {
		t.Fatal("server ID not found in reply")
	}

	if !bytes.Equal(want, got) {
		t.Fatalf("unexpected server ID bytes:\n- want: %v\n-  got: %v", want, got)
	}
}

// TestServeCreateResponseSenderWithCorrectParameters verifies that a new ResponseSender
// gets appropriate transaction ID, client ID, and server ID values copied into
// it before a Handler is invoked.
func TestServeCreateResponseSenderWithCorrectParameters(t *testing.T) {
	txID := [3]byte{0, 1, 2}
	duid := dhcp6opts.NewDUIDLL(1, []byte{0, 1, 0, 1, 0, 1})

	p := &dhcp6.Packet{
		MessageType:   dhcp6.MessageTypeSolicit,
		TransactionID: txID,
		Options:       make(dhcp6.Options),
	}
	p.Options.Add(dhcp6.OptionClientID, duid)

	pb, err := p.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	r := &testMessage{}
	r.b.Write(pb)

	// Do not expect a reply, but do some validation to ensure that Serve
	// sets up appropriate Request and ResponseSender values from an input request
	_, _, err = testServe(r, nil, false, func(w ResponseSender, r *Request) {
		if want, got := txID[:], r.TransactionID[:]; !bytes.Equal(want, got) {
			t.Fatalf("unexpected transaction ID:\n- want: %v\n-  got: %v", want, got)
		}

		cID, err := dhcp6opts.GetClientID(w.Options())
		if err != nil {
			t.Fatal("ResponseSender options did not contain client ID")
		}
		if want, got := duid, cID; !reflect.DeepEqual(want, got) {
			t.Fatalf("unexpected client ID bytes:\n- want: %v\n-  got: %v", want, got)
		}

		if sID, err := dhcp6opts.GetServerID(w.Options()); err != nil || sID == nil {
			t.Fatal("ResponseSender options did not contain server ID")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestServeIgnoreWrongCMIfIndex verifies that Serve will ignore incoming
// connections with an incorrect IPv6 control message interface index.
func TestServeIgnoreWrongCMIfIndex(t *testing.T) {
	// Wrong interface index in control message
	r := &testMessage{
		cm: &ipv6.ControlMessage{
			IfIndex: -1,
		},
	}
	r.b.Write([]byte{0, 0, 0, 0})

	s := &Server{
		Iface: &net.Interface{
			Index: 0,
		},
	}

	// Expect no reply at all
	w, _, err := testServe(r, s, false, func(w ResponseSender, r *Request) {})
	if err != nil {
		t.Fatal(err)
	}

	if l := w.b.Len(); l > 0 {
		t.Fatalf("reply should be empty, but got length: %d", l)
	}
	if cm := w.cm; cm != nil {
		t.Fatalf("control message should be nil, but got: %v", cm)
	}
	if addr := w.addr; addr != nil {
		t.Fatalf("address should be nil, but got: %v", addr)
	}
}

// TestServeIgnoreInvalidPacket verifies that Serve will ignore invalid
// request packets.
func TestServeIgnoreInvalidPacket(t *testing.T) {
	// Packet too short to be valid
	r := &testMessage{}
	r.b.Write([]byte{0, 0, 0})

	// Expect no reply at all
	w, _, err := testServe(r, nil, false, func(w ResponseSender, r *Request) {})
	if err != nil {
		t.Fatal(err)
	}

	if l := w.b.Len(); l > 0 {
		t.Fatalf("reply should be empty, but got length: %d", l)
	}
	if cm := w.cm; cm != nil {
		t.Fatalf("control message should be nil, but got: %v", cm)
	}
	if addr := w.addr; addr != nil {
		t.Fatalf("address should be nil, but got: %v", addr)
	}
}

// TestServeIgnoreBadMessageType verifies that Serve will ignore request
// packets with invalid message types.
func TestServeIgnoreBadMessageType(t *testing.T) {
	// Message types not known
	badMT := []byte{0, 22}
	for _, mt := range badMT {
		r := &testMessage{}
		r.b.Write([]byte{mt, 0, 0, 0})

		// Expect no reply at all
		w, _, err := testServe(r, nil, false, func(w ResponseSender, r *Request) {})
		if err != nil {
			t.Fatal(err)
		}

		if l := w.b.Len(); l > 0 {
			t.Fatalf("reply should be empty, but got length: %d", l)
		}
		if cm := w.cm; cm != nil {
			t.Fatalf("control message should be nil, but got: %v", cm)
		}
		if addr := w.addr; addr != nil {
			t.Fatalf("address should be nil, but got: %v", addr)
		}
	}
}

// TestServeOK verifies that Serve correctly handles an incoming request and
// all of its options, and replies with expected values.
func TestServeOK(t *testing.T) {
	txID := [3]byte{0, 1, 2}
	duid := dhcp6opts.NewDUIDLL(1, []byte{0, 1, 0, 1, 0, 1})

	// Perform an entire Solicit transaction
	p := &dhcp6.Packet{
		MessageType:   dhcp6.MessageTypeSolicit,
		TransactionID: txID,
		Options:       make(dhcp6.Options),
	}
	p.Options.Add(dhcp6.OptionClientID, duid)

	pb, err := p.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	// Send from a different mock IP
	r := &testMessage{
		addr: &net.UDPAddr{
			IP: net.ParseIP("::2"),
		},
	}
	r.b.Write(pb)

	// Expect these option values set by server
	var preference dhcp6opts.Preference = 255
	sCode := dhcp6.StatusSuccess
	sMsg := "success"

	// Expect Advertise reply with several options added by server
	mt := dhcp6.MessageTypeAdvertise
	w, _, err := testServe(r, nil, true, func(w ResponseSender, r *Request) {
		w.Options().Add(dhcp6.OptionPreference, preference)
		w.Options().Add(dhcp6.OptionStatusCode, dhcp6opts.NewStatusCode(sCode, sMsg))

		w.Send(mt)
	})
	if err != nil {
		t.Fatal(err)
	}

	// For now, outgoing control message is always nil
	if w.cm != nil {
		t.Fatal("control message is never set on outgoing reply")
	}
	if want, got := r.addr, w.addr; want != got {
		t.Fatalf("unexpected client address: %v != %v", want, got)
	}

	wp := new(dhcp6.Packet)
	if err := wp.UnmarshalBinary(w.b.Bytes()); err != nil {
		t.Fatal(err)
	}

	if want, got := mt, wp.MessageType; want != got {
		t.Fatalf("unexpected message type: %v != %v", want, got)
	}

	if want, got := txID[:], wp.TransactionID[:]; !bytes.Equal(want, got) {
		t.Fatalf("unexpected transaction ID:\n- want: %v\n-  got: %v", want, got)
	}

	cID, err := dhcp6opts.GetClientID(wp.Options)
	if err != nil {
		t.Fatalf("response options did not contain client ID: %v", err)
	}
	if want, got := duid, cID; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected client ID bytes:\n- want: %v\n-  got: %v", want, got)
	}
	if sID, err := dhcp6opts.GetServerID(wp.Options); err != nil || sID == nil {
		t.Fatal("ResponseSender options did not contain server ID")
	}

	pr, err := dhcp6opts.GetPreference(wp.Options)
	if err != nil {
		t.Fatal("response Options did not contain preference")
	}
	if want, got := preference, pr; want != got {
		t.Fatalf("unexpected preference value: %v != %v", want, got)
	}

	st, err := dhcp6opts.GetStatusCode(wp.Options)
	if err != nil {
		t.Fatal("response Options did not contain status code")
	}
	if want, got := sCode, st.Code; want != got {
		t.Fatalf("unexpected status code value: %v != %v", want, got)
	}
	if want, got := sMsg, st.Message; want != got {
		t.Fatalf("unexpected status code meesage: %q != %q", want, got)
	}
}

// testServe performs a single transaction using the input message, server
// configuration, whether or not a reply is expected, and a closure which
// acts as a HandlerFunc.
func testServe(r *testMessage, s *Server, expectReply bool, fn func(w ResponseSender, r *Request)) (*testMessage, *recordIPv6PacketConn, error) {
	// If caller doesn't specify a testMessage or client address
	// for it, configure it for them
	if r == nil {
		r = &testMessage{}
	}
	if r.addr == nil {
		r.addr = &net.UDPAddr{
			IP: net.ParseIP("::1"),
		}
	}

	// If caller doesn't specify Server value, configure it for
	// them using input function as HandlerFunc.
	if s == nil {
		s = &Server{}
	}
	if s.Iface == nil {
		s.Iface = &net.Interface{
			Name:  "foo0",
			Index: 0,
		}
	}
	s.Handler = HandlerFunc(fn)

	// Implements PacketConn to capture request/response
	tc := &testPacketConn{
		r: r,
		w: &testMessage{},

		// Record IPv6 control parameters
		recordIPv6PacketConn: &recordIPv6PacketConn{
			joined: make([]net.Addr, 0),
			left:   make([]net.Addr, 0),
			flags:  make(map[ipv6.ControlFlags]bool),
		},
	}

	// Perform a single read and possibly write before returning
	// an error on second read to close server
	c := &oneReadPacketConn{
		PacketConn: tc,

		readDoneC:  make(chan struct{}, 0),
		writeDoneC: make(chan struct{}, 0),
	}

	// If no reply is expected, this channel will never be closed,
	// and should be closed immediately
	if !expectReply {
		close(c.writeDoneC)
	}

	// Handle request
	err := s.Serve(c)

	// Wait for read and write to complete
	<-c.readDoneC
	<-c.writeDoneC

	// Return written values, IPv6 control parameters
	return tc.w, tc.recordIPv6PacketConn, err
}

// oneReadPacketConn allows a single read and possibly write transaction using
// the embedded PacketConn before issuing errClosing to close the server.
type oneReadPacketConn struct {
	PacketConn

	err    error
	txDone bool

	readDoneC  chan struct{}
	writeDoneC chan struct{}
}

// ReadFrom reads input bytes using the underlying PacketConn only once.  Once
// the read is completed, readDoneC will close and stop blocking.  Any
// further ReadFrom calls result in errClosing.
func (c *oneReadPacketConn) ReadFrom(b []byte) (int, *ipv6.ControlMessage, net.Addr, error) {
	if c.txDone {
		return 0, nil, nil, errClosing
	}
	c.txDone = true

	n, cm, addr, err := c.PacketConn.ReadFrom(b)
	close(c.readDoneC)
	return n, cm, addr, err
}

// WriteTo writes input bytes and IPv6 control parameters to the underlying
// PacketConn.  Once the write is completed, writeDoneC will close and
// stop blocking.
func (c *oneReadPacketConn) WriteTo(b []byte, cm *ipv6.ControlMessage, dst net.Addr) (int, error) {
	n, err := c.PacketConn.WriteTo(b, cm, dst)
	close(c.writeDoneC)
	return n, err
}

// testPacketConn captures client requests, server responses, and IPv6
// control parameters set by the server.
type testPacketConn struct {
	r *testMessage
	w *testMessage

	*recordIPv6PacketConn
}

// testMessage captures data from a request or response.
type testMessage struct {
	b    bytes.Buffer
	cm   *ipv6.ControlMessage
	addr net.Addr
}

// ReadFrom returns data from a preconfigured testMessage containing bytes,
// an IPv6 control message, and a client address.
func (c *testPacketConn) ReadFrom(b []byte) (int, *ipv6.ControlMessage, net.Addr, error) {
	n, err := c.r.b.Read(b)
	return n, c.r.cm, c.r.addr, err
}

// WriteTo writes data to a testMessage containing bytes, an IPv6 control
// message, and a client address.
func (c *testPacketConn) WriteTo(b []byte, cm *ipv6.ControlMessage, dst net.Addr) (int, error) {
	n, err := c.w.b.Write(b)
	c.w.cm = cm
	c.w.addr = dst

	return n, err
}

// recordIPv6PacketConn tracks IPv6 control parameters, such as joined and
// left multicast groups, control flags, and whether or not the connection
// was closed.
type recordIPv6PacketConn struct {
	closed bool
	joined []net.Addr
	left   []net.Addr
	flags  map[ipv6.ControlFlags]bool
}

// Close records that an IPv6 connection was closed.
func (c *recordIPv6PacketConn) Close() error {
	c.closed = true
	return nil
}

// JoinGroup records that a multicast group was joined.
func (c *recordIPv6PacketConn) JoinGroup(ifi *net.Interface, group net.Addr) error {
	c.joined = append(c.joined, group)
	return nil
}

// LeaveGroup records that a multicast group was left.
func (c *recordIPv6PacketConn) LeaveGroup(ifi *net.Interface, group net.Addr) error {
	c.left = append(c.left, group)
	return nil
}

// SetControlMessage records that an IPv6 control message was set.
func (c *recordIPv6PacketConn) SetControlMessage(cf ipv6.ControlFlags, on bool) error {
	c.flags[cf] = on
	return nil
}
