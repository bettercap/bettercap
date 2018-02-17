package network

import (
	"fmt"
	"net"
	"regexp"

	"github.com/evilsocket/bettercap-ng/core"
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
	out, err := core.Exec(airPortPath, []string{iface, "--channel", fmt.Sprintf("%d", channel)})
	if err != nil {
		return err
	} else if out != "" {
		return fmt.Errorf("Unexpected output while setting interface %s to channel %d: %s", iface, channel, out)
	}
	return nil
}
