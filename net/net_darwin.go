package net

import (
	"net"
	"regexp"
)

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
