package network

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/google/gopacket/pcap"
)

// only matches gateway lines
var IPv4RouteParser = regexp.MustCompile("^.+\\s+.+\\s+\\d+\\s+([0-9\\.]+/\\d+)\\s+\\d+\\s+([0-9\\.]+).*$")
var IPv4RouteTokens = 3
var IPv4RouteCmd = "netsh"
var IPv4RouteCmdOpts = []string{"interface", "ipv4", "show", "route"}

func IPv4RouteIsGateway(ifname string, tokens []string, f func(gateway string) (*Endpoint, error)) (*Endpoint, error) {
	// TODO check if the subnet is the same as iface ?
	// subnet := tokens[1]
	gateway := tokens[2]
	return f(gateway)
}

/*
 * net.Interface does not have the correct name on Windows and pcap.Interface
 * does not have the hardware address for some reason ... so this is what I
 * had to do in Windows ... tnx Microsoft <3
 *
 * FIXME: Just to be clear *THIS IS SHIT*. Please someone test, find a more
 * elegant solution and refactor ... i'm seriously tired of this.
 */

func areTheSame(iface net.Interface, piface pcap.Interface) bool {
	if addrs, err := iface.Addrs(); err == nil {
		for _, ia := range addrs {
			for _, ib := range piface.Addresses {
				if ia.String() == ib.IP.String() || strings.HasPrefix(ia.String(), ib.IP.String()) {
					return true
				}
			}
		}
	}
	return false
}

func getInterfaceName(iface net.Interface) string {
	devs, err := pcap.FindAllDevs()
	if err != nil {
		return iface.Name
	}

	for _, dev := range devs {
		if areTheSame(iface, dev) {
			return dev.Name
		}
	}

	return iface.Name
}

func SetInterfaceChannel(iface string, channel int) error {
	return fmt.Errorf("Windows does not support WiFi channel hopping.")
}

func GetSupportedFrequencies(iface string) ([]int, error) {
	freqs := make([]int, 0)
	return freqs, fmt.Errorf("Windows does not support WiFi channel hopping.")
}
