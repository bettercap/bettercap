package net_sniff

import (
	"fmt"
	"net"

	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/tui"
	"github.com/google/gopacket/layers"
)

func vIP(ip net.IP) string {
	if session.I.Interface.IP.Equal(ip) {
		return tui.Dim("local")
	} else if session.I.Gateway.IP.Equal(ip) {
		return "gateway"
	}

	address := ip.String()
	host := session.I.Lan.GetByIp(address)
	if host != nil {
		if host.Hostname != "" {
			return host.Hostname
		}
	}

	return address
}

func vPort(p interface{}) string {
	sp := fmt.Sprintf("%d", p)
	if tcp, ok := p.(layers.TCPPort); ok {
		if name, found := layers.TCPPortNames[tcp]; found {
			sp = tui.Yellow(name)
		}
	} else if udp, ok := p.(layers.UDPPort); ok {
		if name, found := layers.UDPPortNames[udp]; found {
			sp = tui.Yellow(name)
		}
	}

	return sp
}

var maxUrlSize = 80

func vURL(u string) string {
	ul := len(u)
	if ul > maxUrlSize {
		u = fmt.Sprintf("%s...", u[0:maxUrlSize-3])
	}
	return u
}
