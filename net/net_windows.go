package net

import (
	"regexp"
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
