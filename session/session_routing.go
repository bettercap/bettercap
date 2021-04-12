package session

import (
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/routing"
	"github.com/evilsocket/islazy/log"
	"time"
)

type gateway struct {
	IP  string `json:"ip"`
	MAC string `json:"mac"`
}

type GatewayChange struct {
	Type string  `json:"type"`
	Prev gateway `json:"prev"`
	New  gateway `json:"new"`
}

func (s *Session) routeMon() {
	var err error
	var gw4 *network.Endpoint
	var gwIP6, gwMAC6 string

	s.Events.Log(log.INFO, "gateway monitor started ...")

	if gw4 = s.Gateway; gw4 == nil {
		gw4 = &network.Endpoint{}
	}

	gwIP6, err = routing.Gateway(routing.IPv6, s.Interface.Name())
	if err != nil {
		s.Events.Log(log.ERROR, "error getting ipv6 gateway: %v", err)
	} else if gwIP6 != "" {
		gwMAC6, err = network.ArpLookup(s.Interface.Name(), gwIP6, true)
		if err != nil {
			s.Events.Log(log.DEBUG, "error getting %s ipv6 gateway mac: %v", gwIP6, err)
		}
	}

	for {
		s.Events.Log(log.DEBUG, "[gw] ipv4=%s(%s) ipv6=%s(%s)", gw4.IP, gw4.HwAddress, gwIP6, gwMAC6)

		time.Sleep(5 * time.Second)

		gw4now, err := network.FindGateway(s.Interface)
		if gw4now == nil {
			gw4now = &network.Endpoint{}
		}

		if err != nil {
			s.Events.Log(log.ERROR, "error getting ipv4 gateway: %v", err)
		} else {
			if gw4now.IpAddress != gw4.IpAddress || gw4now.HwAddress != gw4.HwAddress {
				s.Events.Add("gateway.change", GatewayChange{
					Type: string(routing.IPv4),
					Prev: gateway{
						IP:  gw4.IpAddress,
						MAC: gw4.HwAddress,
					},
					New: gateway{
						IP:  gw4now.IpAddress,
						MAC: gw4now.HwAddress,
					},
				})
			}
		}

		gw4 = gw4now

		gwMAC6now := ""
		gwIP6now, err := routing.Gateway(routing.IPv6, s.Interface.Name())
		if err != nil {
			s.Events.Log(log.ERROR, "error getting ipv6 gateway: %v", err)
		} else if gwIP6now != "" {
			gwMAC6now, err = network.ArpLookup(s.Interface.Name(), gwIP6now, true)
			if err != nil {
				s.Events.Log(log.DEBUG, "error getting %s ipv6 gateway mac: %v", gwIP6now, err)
			}
		}

		if gwIP6now != gwIP6 || gwMAC6now != gwMAC6 {
			s.Events.Add("gateway.change", GatewayChange{
				Type: string(routing.IPv6),
				Prev: gateway{
					IP:  gwIP6,
					MAC: gwMAC6,
				},
				New: gateway{
					IP:  gwIP6now,
					MAC: gwMAC6now,
				},
			})
		}

		gwIP6 = gwIP6now
		gwMAC6 = gwMAC6now
	}
}
