package network

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/bettercap/bettercap/core"
)

// only matches gateway lines
var IPv4RouteParser = regexp.MustCompile(`^(default|[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)\svia\s([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)\sdev\s(\S+).*$`)
var IPv4RouteTokens = 4
var IPv4RouteCmd = "ip"
var IPv4RouteCmdOpts = []string{"route"}
var WiFiFreqParser = regexp.MustCompile(`^\s+Channel.([0-9]+)\s+:\s+([0-9\.]+)\s+GHz.*$`)

func IPv4RouteIsGateway(ifname string, tokens []string, f func(gateway string) (*Endpoint, error)) (*Endpoint, error) {
	ifname2 := tokens[3]

	if ifname == ifname2 {
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

	out, err := core.Exec("iwconfig", []string{iface, "channel", fmt.Sprintf("%d", channel)})
	if err != nil {
		return err
	} else if out != "" {
		return fmt.Errorf("Unexpected output while setting interface %s to channel %d: %s", iface, channel, out)
	}

	SetInterfaceCurrentChannel(iface, channel)
	return nil
}

func processSupportedFrequencies(output string, err error) ([]int, error) {
	freqs := make([]int, 0)
	if err != nil {
		return freqs, err
	} else if output != "" {
		scanner := bufio.NewScanner(strings.NewReader(output))
		for scanner.Scan() {
			line := scanner.Text()
			matches := WiFiFreqParser.FindStringSubmatch(line)
			if len(matches) == 3 {
				if freq, err := strconv.ParseFloat(matches[2], 64); err == nil {
					freqs = append(freqs, int(freq*1000))
				}
			}
		}
	}
	return freqs, nil
}

func GetSupportedFrequencies(iface string) ([]int, error) {
	out, err := core.Exec("iwlist", []string{iface, "freq"})
	return processSupportedFrequencies(out, err)
}
