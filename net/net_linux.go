package net

import "regexp"

// only matches gateway lines
var IPv4RouteParser = regexp.MustCompile("^(default|[0-9\\.]+)\\svia\\s([0-9\\.]+)\\sdev\\s(\\w+)\\s.*$")
var IPv4RouteTokens = 4
var IPv4RouteCmd = "ip"
var IPv4RouteCmdOpts = []string{"route"}

func IPv4RouteIsGateway(ifname string, tokens []string, f func(gateway string) (*Endpoint, error)) (*Endpoint, error) {
	ifname2 := tokens[3]

	if ifname == ifname2 {
		gateway := tokens[2]
		return f(gateway)
	}

	return nil, nil
}
