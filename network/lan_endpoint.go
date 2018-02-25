package network

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/bettercap/bettercap/core"
)

type OnHostResolvedCallback func(e *Endpoint)

type Endpoint struct {
	Index            int                    `json:"-"`
	IP               net.IP                 `json:"-"`
	Net              *net.IPNet             `json:"-"`
	IPv6             net.IP                 `json:"-"`
	HW               net.HardwareAddr       `json:"-"`
	IpAddress        string                 `json:"ipv4"`
	Ip6Address       string                 `json:"ipv6"`
	SubnetBits       uint32                 `json:"-"`
	IpAddressUint32  uint32                 `json:"-"`
	HwAddress        string                 `json:"mac"`
	Hostname         string                 `json:"hostname"`
	Alias            string                 `json:"alias"`
	Vendor           string                 `json:"vendor"`
	ResolvedCallback OnHostResolvedCallback `json:"-"`
	FirstSeen        time.Time              `json:"first_seen"`
	LastSeen         time.Time              `json:"last_seen"`
	Meta             *Meta                  `json:"meta"`
}

func NewEndpointNoResolve(ip, mac, name string, bits uint32) *Endpoint {
	mac = NormalizeMac(mac)
	hw, _ := net.ParseMAC(mac)
	now := time.Now()

	e := &Endpoint{
		IP:               nil,
		IpAddress:        ip,
		IpAddressUint32:  0,
		Net:              nil,
		HW:               hw,
		SubnetBits:       bits,
		HwAddress:        mac,
		Hostname:         name,
		Vendor:           OuiLookup(mac),
		ResolvedCallback: nil,
		FirstSeen:        now,
		LastSeen:         now,
		Meta:             NewMeta(),
	}

	e.SetIP(ip)
	e.SetBits(bits)

	return e
}

func NewEndpoint(ip, mac string) *Endpoint {
	e := NewEndpointNoResolve(ip, mac, "", 0)
	// start resolver goroutine
	go func() {
		if names, err := net.LookupAddr(e.IpAddress); err == nil && len(names) > 0 {
			e.Hostname = names[0]
			if e.ResolvedCallback != nil {
				e.ResolvedCallback(e)
			}
		}
	}()

	return e
}

func NewEndpointWithAlias(ip, mac, alias string) *Endpoint {
	e := NewEndpoint(ip, mac)
	e.Alias = alias
	return e
}

func ip2int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func (t *Endpoint) SetNetwork(netw string) {
	parts := strings.Split(netw, "/")
	address := parts[0]
	bits, _ := strconv.Atoi(parts[1])

	t.SetIP(address)
	t.SetBits(uint32(bits))
}

func (t *Endpoint) SetIPv6(netw string) {
	parts := strings.SplitN(netw, "/", 2)
	address := parts[0]
	// bits, _ := strconv.Atoi(parts[1])

	t.IPv6 = net.ParseIP(address)
	if t.IPv6 != nil {
		t.Ip6Address = t.IPv6.String()
	}
}

func (t *Endpoint) SetIP(ip string) {
	addr := net.ParseIP(ip)
	t.IP = addr
	t.IpAddress = ip
	t.IpAddressUint32 = ip2int(addr)
}

func (t *Endpoint) SetBits(bits uint32) {
	t.SubnetBits = bits
	_, netw, _ := net.ParseCIDR(t.CIDR())
	t.Net = netw
}

func (t *Endpoint) Name() string {
	return t.Hostname
}

func (t *Endpoint) CIDR() string {
	shift := 32 - t.SubnetBits
	address := t.IpAddressUint32
	ip := make(net.IP, 4)

	binary.BigEndian.PutUint32(ip, (address>>shift)<<shift)

	return fmt.Sprintf("%s/%d", ip.String(), t.SubnetBits)
}

func (t *Endpoint) IsMonitor() bool {
	return t.IpAddress == MonitorModeAddress
}

func (t *Endpoint) String() string {
	if t.HwAddress == "" {
		return t.IpAddress
	} else if t.Vendor == "" {
		return fmt.Sprintf("%s : %s", t.IpAddress, t.HwAddress)
	} else if t.Hostname == "" {
		return fmt.Sprintf("%s : %s ( %s )", t.IpAddress, t.HwAddress, t.Vendor)
	}

	return fmt.Sprintf("%s : %s ( %s ) - %s", t.IpAddress, t.HwAddress, t.Vendor, core.Bold(t.Hostname))
}
