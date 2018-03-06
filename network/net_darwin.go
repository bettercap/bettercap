package network

import (
	"fmt"
	"net"
	"regexp"

	"github.com/bettercap/bettercap/core"
)

const airPortPath = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"

var IPv4RouteParser = regexp.MustCompile("^([a-z]+)+\\s+(\\d+\\.+\\d+.\\d.+\\d)+\\s+([a-zA-z]+)+\\s+(\\d+)+\\s+(\\d+)+\\s+([a-zA-Z]+\\d+)$")
var IPv4RouteTokens = 7
var IPv4RouteCmd = "netstat"
var IPv4RouteCmdOpts = []string{"-n", "-r"}

func IPv4RouteIsGateway(ifname string, tokens []string, f func(gateway string) (*Endpoint, error)) (*Endpoint, error) {
	ifname2 := tokens[6]
	flags := tokens[3]

	if ifname == ifname2 && flags == "UGSc" {
		gateway := tokens[2]
		return f(gateway)
	}

	return nil, nil
}

// see Windows version to understand why ....
func getInterfaceName(iface net.Interface) string {
	return iface.Name
}

func SetInterfaceChannel(iface string, channel int) error {
	_, err := core.Exec(airPortPath, []string{iface, fmt.Sprintf("-c%d", channel)})
	return err
}

//! TODO Get the list of the available frequencies supported by the network card
func GetSupportedFrequencies(iface string) ([]int, error) {
	freqs := []int{2412, 2417, 2422, 2427, 2432, 2437, 2442, 2447, 2452, 2457, 2462, 2467, 2472, 2484}
	return freqs, nil
}
