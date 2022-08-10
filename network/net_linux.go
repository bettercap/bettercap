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

	if core.HasBinary("iw") {
		// Debug("SetInterfaceChannel(%s, %d) iw based", iface, channel)
		out, err := core.Exec("iw", []string{"dev", iface, "set", "channel", fmt.Sprintf("%d", channel)})
		if err != nil {
			return fmt.Errorf("iw: out=%s err=%s", out, err)
		} else if out != "" {
			return fmt.Errorf("Unexpected output while setting interface %s to channel %d: %s", iface, channel, out)
		}
	} else if core.HasBinary("iwconfig") {
		// Debug("SetInterfaceChannel(%s, %d) iwconfig based")
		out, err := core.Exec("iwconfig", []string{iface, "channel", fmt.Sprintf("%d", channel)})
		if err != nil {
			return fmt.Errorf("iwconfig: out=%s err=%s", out, err)
		} else if out != "" {
			return fmt.Errorf("Unexpected output while setting interface %s to channel %d: %s", iface, channel, out)
		}
	} else {
		return fmt.Errorf("no iw or iwconfig binaries found in $PATH")
	}

	SetInterfaceCurrentChannel(iface, channel)
	return nil
}

var iwlistFreqParser = regexp.MustCompile(`^\s+Channel.([0-9]+)\s+:\s+([0-9\.]+)\s+GHz.*$`)

func iwlistSupportedFrequencies(iface string) ([]int, error) {
	out, err := core.Exec("iwlist", []string{iface, "freq"})
	if err != nil {
		return nil, err
	}

	freqs := make([]int, 0)
	if out != "" {
		scanner := bufio.NewScanner(strings.NewReader(out))
		for scanner.Scan() {
			line := scanner.Text()
			matches := iwlistFreqParser.FindStringSubmatch(line)
			if len(matches) == 3 {
				if freq, err := strconv.ParseFloat(matches[2], 64); err == nil {
					freqs = append(freqs, int(freq*1000))
				}
			}
		}
	}

	return freqs, nil
}

var iwPhyParser = regexp.MustCompile(`^\s*wiphy\s+(\d+)$`)
var iwFreqParser = regexp.MustCompile(`^\s+\*\s+(\d+)\s+MHz.+dBm.+$`)

func iwSupportedFrequencies(iface string) ([]int, error) {
	// first determine phy index
	out, err := core.Exec("iw", []string{iface, "info"})
	if err != nil {
		return nil, fmt.Errorf("error getting %s phy index: %v", iface, err)
	}

	phy := int64(-1)
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		matches := iwPhyParser.FindStringSubmatch(line)
		if len(matches) == 2 {
			if phy, err = strconv.ParseInt(matches[1], 10, 32); err != nil {
				return nil, fmt.Errorf("error parsing %s phy index: %v (line: %s)", iface, err, line)
			}
		}
	}

	if phy == -1 {
		return nil, fmt.Errorf("could not find %s phy index", iface)
	}

	// then get phyN info
	phyName := fmt.Sprintf("phy%d", phy)
	out, err = core.Exec("iw", []string{phyName, "info"})
	if err != nil {
		return nil, fmt.Errorf("error getting %s (%s) info: %v", phyName, iface, err)
	}

	freqs := []int{}
	scanner = bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		matches := iwFreqParser.FindStringSubmatch(line)
		if len(matches) == 2 {
			if freq, err := strconv.ParseInt(matches[1], 10, 64); err != nil {
				return nil, fmt.Errorf("error parsing %s freq: %v (line: %s)", iface, err, line)
			} else {
				freqs = append(freqs, int(freq))
			}
		}
	}

	return freqs, nil
}

func GetSupportedFrequencies(iface string) ([]int, error) {
	// give priority to iwlist because of https://github.com/bettercap/bettercap/issues/881
	if core.HasBinary("iwlist") {
		return iwlistSupportedFrequencies(iface)
	} else if core.HasBinary("iw") {
		return iwSupportedFrequencies(iface)
	}

	return nil, fmt.Errorf("no iw or iwlist binaries found in $PATH")
}
