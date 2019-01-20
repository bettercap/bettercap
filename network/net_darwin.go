package network

import (
	"fmt"
	"net"
	"regexp"
	"strconv"

	"github.com/bettercap/bettercap/core"

	"github.com/evilsocket/islazy/str"
)

const airPortPath = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"

var IPv4RouteParser = regexp.MustCompile(`^([a-z]+)+\s+(\d+\.+\d+.\d.+\d)+\s+([a-zA-z]+)+\s+(\d+)+\s+(\d+)+\s+([a-zA-Z]+\d+)$`)
var IPv4RouteTokens = 7
var IPv4RouteCmd = "netstat"
var IPv4RouteCmdOpts = []string{"-n", "-r"}
var WiFiChannelParser = regexp.MustCompile(`(?m)^.*Supported Channels: (.*)$`)

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
	curr := GetInterfaceChannel(iface)
	// the interface is already on this channel
	if curr == channel {
		return nil
	}

	_, err := core.Exec(airPortPath, []string{iface, fmt.Sprintf("-c%d", channel)})
	if err != nil {
		return err
	}

	SetInterfaceCurrentChannel(iface, channel)
	return nil
}

func getFrequenciesFromChannels(output string) ([]int, error) {
	freqs := make([]int, 0)
	if output != "" {
		if matches := WiFiChannelParser.FindStringSubmatch(output); len(matches) == 2 {
			for _, channel := range str.Comma(matches[1]) {
				if channel, err := strconv.Atoi(channel); err == nil {
					freqs = append(freqs, Dot11Chan2Freq(channel))
				}
			}
		}
	}
	return freqs, nil
}

func GetSupportedFrequencies(iface string) ([]int, error) {
	out, err := core.Exec("system_profiler", []string{"SPAirPortDataType"})
	if err != nil {
		return nil, err
	}
	return getFrequenciesFromChannels(out)
}
