package syn_scan

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/evilsocket/islazy/str"
	"github.com/malfunkt/iprange"
)

func (mod *SynScanner) parseTargets(arg string) error {
	if strings.Contains(arg, ":") {
		// parse as IPv6 address
		if ip := net.ParseIP(arg); ip == nil {
			return fmt.Errorf("error while parsing IPv6 '%s'", arg)
		} else {
			mod.addresses = []net.IP{ip}
		}
	} else {
		if list, err := iprange.Parse(arg); err != nil {
			return fmt.Errorf("error while parsing IP range '%s': %s", arg, err)
		} else {
			mod.addresses = list.Expand()
		}
	}

	return nil
}

func (mod *SynScanner) parsePorts(args []string) (err error) {
	argc := len(args)
	mod.stats.totProbes = 0
	mod.stats.doneProbes = 0
	mod.startPort = 1
	mod.endPort = 65535

	if argc > 1 && str.Trim(args[1]) != "" {
		if mod.startPort, err = strconv.Atoi(str.Trim(args[1])); err != nil {
			return fmt.Errorf("invalid start port %s: %s", args[1], err)
		} else if mod.startPort > 65535 {
			mod.startPort = 65535
		}
		mod.endPort = mod.startPort
	}

	if argc > 2 && str.Trim(args[2]) != "" {
		if mod.endPort, err = strconv.Atoi(str.Trim(args[2])); err != nil {
			return fmt.Errorf("invalid end port %s: %s", args[2], err)
		}
	}

	if mod.endPort < mod.startPort {
		return fmt.Errorf("end port %d is greater than start port %d", mod.endPort, mod.startPort)
	}

	return
}
