package net

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/evilsocket/bettercap-ng/core"
)

type Endpoint struct {
	IP              net.IP
	HW              net.HardwareAddr
	IpAddress       string
	SubnetBits      uint32
	IpAddressUint32 uint32
	HwAddress       string
	Hostname        string
	Vendor          string
}

type OnHostResolvedAction func(e *Endpoint)

func NewEndpointNoResolve(ip, mac, name string, bits uint32) *Endpoint {
	hw, err := net.ParseMAC(mac)
	if err != nil {
		panic(err)
	}

	e := &Endpoint{
		IP:              net.ParseIP(ip),
		HW:              hw,
		IpAddress:       ip,
		SubnetBits:      bits,
		IpAddressUint32: binary.BigEndian.Uint32(net.ParseIP(ip)[12:16]),
		HwAddress:       mac,
		Hostname:        name,
		Vendor:          OuiLookup(mac),
	}

	return e
}

func NewEndpoint(ip, mac string) *Endpoint {
	e := NewEndpointNoResolve(ip, mac, "", 0)

	// start resolver goroutine
	go func() {
		if names, err := net.LookupAddr(e.IpAddress); err == nil {
			e.Hostname = names[0]
			log.Debugf("Endpoint %s is now known as %s\n", e.IpAddress, core.Green(e.Hostname))
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
