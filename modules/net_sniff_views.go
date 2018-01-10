package modules

import (
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket/layers"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/session"
)

func vTime(t time.Time) string {
	return t.Format("15:04:05")
}

func vIP(ip net.IP) string {
	if session.I.Interface.IP.Equal(ip) {
		return core.Dim("local")
	} else if session.I.Gateway.IP.Equal(ip) {
		return "gateway"
	}

	address := ip.String()
	host, found := session.I.Targets.Targets[address]

	if found == true {
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
			sp = core.Yellow(name)
		}
	} else if udp, ok := p.(layers.UDPPort); ok {
		if name, found := layers.UDPPortNames[udp]; found {
			sp = core.Yellow(name)
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
