// +build linux

package raw

import (
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/net/bpf"
	"golang.org/x/sys/unix"
)

var (
	// Must implement net.PacketConn at compile-time.
	_ net.PacketConn = &packetConn{}
)

// packetConn is the Linux-specific implementation of net.PacketConn for this
// package.
type packetConn struct {
	ifi *net.Interface
	s   socket
	pbe uint16

	// Should timeouts be set at all?
	noTimeouts bool

	// Should stats be accumulated instead of reset on each call?
	noCumulativeStats bool

	// Internal storage for cumulative stats.
	stats Stats

	// Timeouts set via Set{Read,}Deadline, guarded by mutex.
	timeoutMu sync.RWMutex
	rtimeout  time.Time
}

// socket is an interface which enables swapping out socket syscalls for
// testing.
type socket interface {
	Bind(unix.Sockaddr) error
	Close() error
	FD() int
	GetSockopt(level, name int, v unsafe.Pointer, l uintptr) error
	Recvfrom([]byte, int) (int, unix.Sockaddr, error)
	Sendto([]byte, int, unix.Sockaddr) error
	SetSockopt(level, name int, v unsafe.Pointer, l uint32) error
	SetTimeout(time.Duration) error
}

// listenPacket creates a net.PacketConn which can be used to send and receive
// data at the device driver level.
func listenPacket(ifi *net.Interface, proto uint16, cfg Config) (*packetConn, error) {
	// Convert proto to big endian.
	pbe := htons(proto)

	// Enabling overriding the socket type via config.
	typ := unix.SOCK_RAW
	if cfg.LinuxSockDGRAM {
		typ = unix.SOCK_DGRAM
	}

	// Open a packet socket using specified socket and protocol types.
	sock, err := unix.Socket(unix.AF_PACKET, typ, int(pbe))
	if err != nil {
		return nil, err
	}

	// Wrap raw socket in socket interface.
	pc, err := newPacketConn(ifi, &sysSocket{fd: sock}, pbe)
	if err != nil {
		return nil, err
	}

	pc.noTimeouts = cfg.NoTimeouts
	pc.noCumulativeStats = cfg.NoCumulativeStats
	return pc, nil
}

// newPacketConn creates a net.PacketConn using the specified network
// interface, wrapped socket and big endian protocol number.
//
// It is the entry point for tests in this package.
func newPacketConn(ifi *net.Interface, s socket, pbe uint16) (*packetConn, error) {
	// Bind the packet socket to the interface specified by ifi
	// packet(7):
	//   Only the sll_protocol and the sll_ifindex address fields are used for
	//   purposes of binding.
	err := s.Bind(&unix.SockaddrLinklayer{
		Protocol: pbe,
		Ifindex:  ifi.Index,
	})
	if err != nil {
		return nil, err
	}

	return &packetConn{
		ifi: ifi,
		s:   s,
		pbe: pbe,
	}, nil
}

// ReadFrom implements the net.PacketConn.ReadFrom method.
func (p *packetConn) ReadFrom(b []byte) (int, net.Addr, error) {
	p.timeoutMu.Lock()
	deadline := p.rtimeout
	p.timeoutMu.Unlock()

	var (
		// Information returned by unix.Recvfrom.
		n    int
		addr unix.Sockaddr
		err  error

		// Timeout for a single loop iteration.
		timeout = readTimeout
	)

	for {
		if !deadline.IsZero() {
			timeout = deadline.Sub(time.Now())
			if timeout > readTimeout {
				timeout = readTimeout
			}
		}

		// Set a timeout for this iteration if configured to do so.
		if !p.noTimeouts {
			if err := p.s.SetTimeout(timeout); err != nil {
				return 0, nil, err
			}
		}

		// Attempt to receive on socket
		// The recvfrom sycall will NOT be interrupted by closing of the socket
		n, addr, err = p.s.Recvfrom(b, 0)
		switch err {
		case nil:
			// Got data, break this loop shortly.
		case unix.EAGAIN:
			// Hit a timeout, keep looping.
			continue
		default:
			// Return on any other error.
			return n, nil, err
		}

		// Got data, exit the loop.
		break
	}

	// Retrieve hardware address and other information from addr.
	sa, ok := addr.(*unix.SockaddrLinklayer)
	if !ok || sa.Halen < 6 {
		return n, nil, unix.EINVAL
	}

	// Use length specified to convert byte array into a hardware address slice.
	mac := make(net.HardwareAddr, sa.Halen)
	copy(mac, sa.Addr[:])

	// packet(7):
	//   sll_hatype and sll_pkttype are set on received packets for your
	//   information.
	// TODO(mdlayher): determine if similar fields exist and are useful on
	// non-Linux platforms
	return n, &Addr{
		HardwareAddr: mac,
	}, nil
}

// WriteTo implements the net.PacketConn.WriteTo method.
func (p *packetConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	// Ensure correct Addr type.
	a, ok := addr.(*Addr)
	if !ok || a.HardwareAddr == nil || len(a.HardwareAddr) < 6 {
		return 0, unix.EINVAL
	}

	// Convert hardware address back to byte array form.
	var baddr [8]byte
	copy(baddr[:], a.HardwareAddr)

	// Send message on socket to the specified hardware address from addr
	// packet(7):
	//   When you send packets it is enough to specify sll_family, sll_addr,
	//   sll_halen, sll_ifindex, and sll_protocol. The other fields should
	//   be 0.
	// In this case, sll_family is taken care of automatically by unix.
	err := p.s.Sendto(b, 0, &unix.SockaddrLinklayer{
		Ifindex:  p.ifi.Index,
		Halen:    uint8(len(a.HardwareAddr)),
		Addr:     baddr,
		Protocol: p.pbe,
	})
	return len(b), err
}

// Close closes the connection.
func (p *packetConn) Close() error {
	return p.s.Close()
}

// LocalAddr returns the local network address.
func (p *packetConn) LocalAddr() net.Addr {
	return &Addr{
		HardwareAddr: p.ifi.HardwareAddr,
	}
}

// SetDeadline implements the net.PacketConn.SetDeadline method.
func (p *packetConn) SetDeadline(t time.Time) error {
	return p.SetReadDeadline(t)
}

// SetReadDeadline implements the net.PacketConn.SetReadDeadline method.
func (p *packetConn) SetReadDeadline(t time.Time) error {
	p.timeoutMu.Lock()
	p.rtimeout = t
	p.timeoutMu.Unlock()
	return nil
}

// SetWriteDeadline implements the net.PacketConn.SetWriteDeadline method.
func (p *packetConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// SetBPF attaches an assembled BPF program to a raw net.PacketConn.
func (p *packetConn) SetBPF(filter []bpf.RawInstruction) error {
	prog := unix.SockFprog{
		Len:    uint16(len(filter)),
		Filter: (*unix.SockFilter)(unsafe.Pointer(&filter[0])),
	}

	err := p.s.SetSockopt(
		unix.SOL_SOCKET,
		unix.SO_ATTACH_FILTER,
		unsafe.Pointer(&prog),
		uint32(unsafe.Sizeof(prog)),
	)
	if err != nil {
		return os.NewSyscallError("setsockopt", err)
	}

	return nil
}

// SetPromiscuous enables or disables promiscuous mode on the interface, allowing it
// to receive traffic that is not addressed to the interface.
func (p *packetConn) SetPromiscuous(b bool) error {
	mreq := unix.PacketMreq{
		Ifindex: int32(p.ifi.Index),
		Type:    unix.PACKET_MR_PROMISC,
	}

	membership := unix.PACKET_ADD_MEMBERSHIP
	if !b {
		membership = unix.PACKET_DROP_MEMBERSHIP
	}

	return p.s.SetSockopt(unix.SOL_PACKET, membership, unsafe.Pointer(&mreq), unix.SizeofPacketMreq)
}

// Stats retrieves statistics from the Conn.
func (p *packetConn) Stats() (*Stats, error) {
	var s unix.TpacketStats
	if err := p.s.GetSockopt(unix.SOL_PACKET, unix.PACKET_STATISTICS, unsafe.Pointer(&s), unsafe.Sizeof(s)); err != nil {
		return nil, err
	}

	return p.handleStats(s), nil
}

// handleStats handles creation of Stats structures from raw packet socket stats.
func (p *packetConn) handleStats(s unix.TpacketStats) *Stats {
	// Does the caller want instantaneous stats as provided by Linux?  If so,
	// return the structure directly.
	if p.noCumulativeStats {
		return &Stats{
			Packets: uint64(s.Packets),
			Drops:   uint64(s.Drops),
		}
	}

	// The caller wants cumulative stats.  Add stats with the internal stats
	// structure and return a copy of the resulting stats.
	packets := atomic.AddUint64(&p.stats.Packets, uint64(s.Packets))
	drops := atomic.AddUint64(&p.stats.Drops, uint64(s.Drops))

	return &Stats{
		Packets: packets,
		Drops:   drops,
	}
}

// sysSocket is the default socket implementation.  It makes use of
// Linux-specific system calls to handle raw socket functionality.
type sysSocket struct {
	fd int
}

// Method implementations simply invoke the syscall of the same name, but pass
// the file descriptor stored in the sysSocket as the socket to use.
func (s *sysSocket) Bind(sa unix.Sockaddr) error { return unix.Bind(s.fd, sa) }
func (s *sysSocket) Close() error                { return unix.Close(s.fd) }
func (s *sysSocket) FD() int                     { return s.fd }
func (s *sysSocket) GetSockopt(level, name int, v unsafe.Pointer, l uintptr) error {
	_, _, err := unix.Syscall6(unix.SYS_GETSOCKOPT, uintptr(s.fd), uintptr(level), uintptr(name), uintptr(v), uintptr(unsafe.Pointer(&l)), 0)
	if err != 0 {
		return unix.Errno(err)
	}
	return nil
}
func (s *sysSocket) Recvfrom(p []byte, flags int) (int, unix.Sockaddr, error) {
	return unix.Recvfrom(s.fd, p, flags)
}
func (s *sysSocket) Sendto(p []byte, flags int, to unix.Sockaddr) error {
	return unix.Sendto(s.fd, p, flags, to)
}
func (s *sysSocket) SetSockopt(level, name int, v unsafe.Pointer, l uint32) error {
	_, _, err := unix.Syscall6(unix.SYS_SETSOCKOPT, uintptr(s.fd), uintptr(level), uintptr(name), uintptr(v), uintptr(l), 0)
	if err != 0 {
		return unix.Errno(err)
	}
	return nil
}
func (s *sysSocket) SetTimeout(timeout time.Duration) error {
	tv, err := newTimeval(timeout)
	if err != nil {
		return err
	}
	return unix.SetsockoptTimeval(s.fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, tv)
}
