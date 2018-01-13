package net

import (
	"github.com/evilsocket/bettercap-ng/core"
	"strings"
)

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
