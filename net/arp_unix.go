package net

import (
	"github.com/evilsocket/bettercap-ng/core"
	"strings"
)

func ArpUpdate(iface string) (ArpTable, error) {
	arpLock.Lock()
	defer arpLock.Unlock()

	// Signal we parsed the ARP table at least once.
	arpWasParsed = true

	// Run "arp -an" (darwin) or "ip neigh" (linux) and parse the output
	output, err := core.Exec(ArpCmd, ArpCmdOpts)
	if err != nil {
		return arpTable, err
	}

	newTable := make(ArpTable)
	for _, line := range strings.Split(output, "\n") {
		m := ArpTableParser.FindStringSubmatch(line)
		if len(m) == ArpTableTokens {
			address := m[ArpTableTokenIndex[0]]
			mac := m[ArpTableTokenIndex[1]]
			ifname := m[ArpTableTokenIndex[2]]

			if ifname == iface {
				newTable[address] = mac
			}
		}
	}

	arpTable = newTable

	return arpTable, nil
}
