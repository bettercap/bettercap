package net

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/op/go-logging"

	"github.com/bettercap/bettercap/core"
)

type ArpTable map[string]string

var (
	log        = logging.MustGetLogger("mitm")
	arp_parsed = false
	arp_lock   = &sync.Mutex{}
	arp_table  = make(ArpTable)
)

var ArpTableParser = regexp.MustCompile("^[^\\d\\.]+([\\d\\.]+).+\\s+([a-f0-9:]{17}).+\\s+(.+)$")
var ArpTableTokens = 4

func ArpDiff(current, before ArpTable) ArpTable {
	diff := make(ArpTable)
	for ip, mac := range current {
		_, found := before[ip]
		if !found {
			diff[ip] = mac
		}
	}

	return diff
}

func ArpLookup(iface string, address string, refresh bool) (string, error) {
	// Refresh ARP table if first run or if a force refresh has been instructed.
	if ArpParsed() == false || refresh == true {
		if _, err := ArpUpdate(iface); err != nil {
			return "", err
		}
	}

	// Lookup the hardware address of this ip.
	if mac, found := arp_table[address]; found == true {
		return mac, nil
	}

	return "", fmt.Errorf("Could not find mac for %s", address)
}

func ArpParsed() bool {
	arp_lock.Lock()
	defer arp_lock.Unlock()
	return arp_parsed
}

func ArpUpdate(iface string) (ArpTable, error) {
	arp_lock.Lock()
	defer arp_lock.Unlock()

	// Signal we parsed the ARP table at least once.
	arp_parsed = true

	// Run "arp -an" and parse the output.
	output, err := core.Exec("arp", []string{"-a", "-n"})
	if err != nil {
		return arp_table, err
	}

	new_table := make(ArpTable)
	for _, line := range strings.Split(output, "\n") {
		m := ArpTableParser.FindStringSubmatch(line)
		if len(m) == ArpTableTokens {
			address := m[1]
			mac := m[2]
			ifname := m[3]

			if ifname == iface {
				new_table[address] = mac
			}
		}
	}

	arp_table = new_table

	return arp_table, nil
}
