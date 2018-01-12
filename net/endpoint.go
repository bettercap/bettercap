package net

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/evilsocket/bettercap-ng/core"
)

type OnHostResolvedCallback func(e *Endpoint)
type Endpoint struct {
	IP               net.IP                 `json:"-"`
	IPv6             net.IP                 `json:"."`
	HW               net.HardwareAddr       `json:"-"`
	IpAddress        string                 `json:"ipv4"`
	Ip6Address       string                 `json:"ipv6"`
	SubnetBits       uint32                 `json:"-"`
	IpAddressUint32  uint32                 `json:"-"`
	HwAddress        string                 `json:"mac"`
	Hostname         string                 `json:"hostname"`
	Vendor           string                 `json:"vendor"`
	ResolvedCallback OnHostResolvedCallback `json:"-"`
}

func NewEndpointNoResolve(ip, mac, name string, bits uint32) *Endpoint {
	hw, _ := net.ParseMAC(mac)
	e := &Endpoint{
		IP:               net.ParseIP(ip),
		HW:               hw,
		IpAddress:        ip,
		SubnetBits:       bits,
		IpAddressUint32:  binary.BigEndian.Uint32(net.ParseIP(ip)[12:16]),
		HwAddress:        mac,
		Hostname:         name,
		Vendor:           OuiLookup(mac),
		ResolvedCallback: nil,
	}

	return e
}

func NewEndpoint(ip, mac string) *Endpoint {
	e := NewEndpointNoResolve(ip, mac, "", 0)

	// start resolver goroutine
	go func() {
		if names, err := net.LookupAddr(e.IpAddress); err == nil {
			e.Hostname = names[0]
			if e.ResolvedCallback != nil {
				e.ResolvedCallback(e)
			}
		}
	}()

	return e
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

func (t *Endpoint) String() string {
	if t.HwAddress == "" {
		return t.IpAddress
	} else if t.Vendor == "" {
		return fmt.Sprintf("%s : %s", t.IpAddress, t.HwAddress)
	} else if t.Hostname == "" {
		return fmt.Sprintf("%s : %s ( %s )", t.IpAddress, t.HwAddress, t.Vendor)
	} else {
		return fmt.Sprintf("%s : %s ( %s ) - "+core.BOLD+t.Hostname+core.RESET, t.IpAddress, t.HwAddress, t.Vendor)
	}
}
