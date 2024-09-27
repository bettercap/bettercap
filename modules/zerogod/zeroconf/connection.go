// Based on https://github.com/grandcat/zeroconf
package zeroconf

import (
	"fmt"
	"net"

	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

var (
	// Multicast groups used by mDNS
	mdnsGroupIPv4 = net.IPv4(224, 0, 0, 251)
	mdnsGroupIPv6 = net.ParseIP("ff02::fb")

	// mDNS wildcard addresses
	mdnsWildcardAddrIPv4 = &net.UDPAddr{
		IP:   net.ParseIP("224.0.0.0"),
		Port: 5353,
	}
	mdnsWildcardAddrIPv6 = &net.UDPAddr{
		IP: net.ParseIP("ff02::"),
		// IP:   net.ParseIP("fd00::12d3:26e7:48db:e7d"),
		Port: 5353,
	}

	// mDNS endpoint addresses
	ipv4Addr = &net.UDPAddr{
		IP:   mdnsGroupIPv4,
		Port: 5353,
	}
	ipv6Addr = &net.UDPAddr{
		IP:   mdnsGroupIPv6,
		Port: 5353,
	}
)

func joinUdp6Multicast(interfaces []net.Interface) (*ipv6.PacketConn, error) {
	udpConn, err := net.ListenUDP("udp6", mdnsWildcardAddrIPv6)
	if err != nil {
		return nil, err
	}

	// Join multicast groups to receive announcements
	pkConn := ipv6.NewPacketConn(udpConn)
	pkConn.SetControlMessage(ipv6.FlagInterface, true)
	_ = pkConn.SetMulticastHopLimit(255)

	if len(interfaces) == 0 {
		interfaces = listMulticastInterfaces()
	}
	// log.Println("Using multicast interfaces: ", interfaces)

	var failedJoins int
	for _, iface := range interfaces {
		if err := pkConn.JoinGroup(&iface, &net.UDPAddr{IP: mdnsGroupIPv6}); err != nil {
			// log.Println("Udp6 JoinGroup failed for iface ", iface)
			failedJoins++
		}
	}
	if failedJoins == len(interfaces) {
		pkConn.Close()
		return nil, fmt.Errorf("udp6: failed to join any of these interfaces: %v", interfaces)
	}

	return pkConn, nil
}

func joinUdp4Multicast(interfaces []net.Interface) (*ipv4.PacketConn, error) {
	udpConn, err := net.ListenUDP("udp4", mdnsWildcardAddrIPv4)
	if err != nil {
		// log.Printf("[ERR] bonjour: Failed to bind to udp4 mutlicast: %v", err)
		return nil, err
	}

	// Join multicast groups to receive announcements
	pkConn := ipv4.NewPacketConn(udpConn)
	pkConn.SetControlMessage(ipv4.FlagInterface, true)
	_ = pkConn.SetMulticastTTL(255)

	if len(interfaces) == 0 {
		interfaces = listMulticastInterfaces()
	}
	// log.Println("Using multicast interfaces: ", interfaces)

	var failedJoins int
	for _, iface := range interfaces {
		if err := pkConn.JoinGroup(&iface, &net.UDPAddr{IP: mdnsGroupIPv4}); err != nil {
			// log.Println("Udp4 JoinGroup failed for iface ", iface)
			failedJoins++
		}
	}
	if failedJoins == len(interfaces) {
		pkConn.Close()
		return nil, fmt.Errorf("udp4: failed to join any of these interfaces: %v", interfaces)
	}

	return pkConn, nil
}

func listMulticastInterfaces() []net.Interface {
	var interfaces []net.Interface
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, ifi := range ifaces {
		// not up
		if (ifi.Flags & net.FlagUp) == 0 {
			continue
		}
		// localhost
		if (ifi.Flags & net.FlagLoopback) != 0 {
			continue
		}
		// not running
		if (ifi.Flags & net.FlagRunning) == 0 {
			continue
		}
		// vpn and similar
		if (ifi.Flags & net.FlagPointToPoint) != 0 {
			continue
		}

		// at least one ipv4 address assigned
		hasIPv4 := false
		if addresses, _ := ifi.Addrs(); addresses != nil {
			for _, addr := range addresses {
				// ipv4 or ipv4 CIDR
				if ip, ipnet, err := net.ParseCIDR(addr.String()); err == nil {
					if ip.To4() != nil || ipnet.IP.To4() != nil {
						hasIPv4 = true
						break
					}
				} else if ipAddr, ok := addr.(*net.IPAddr); ok {
					if ipAddr.IP.To4() != nil {
						hasIPv4 = true
						break
					}
				}
			}
		}

		if !hasIPv4 {
			continue
		}

		if (ifi.Flags & net.FlagMulticast) > 0 {
			interfaces = append(interfaces, ifi)
		}
	}

	return interfaces
}
