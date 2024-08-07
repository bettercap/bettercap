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

var WiFiChannelParser = regexp.MustCompile(`(?m)^.*Supported Channels: (.*)$`)

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
				re := regexp.MustCompile(`\d+`)
				if channel, err := strconv.Atoi(re.FindString(channel)); err == nil {
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
