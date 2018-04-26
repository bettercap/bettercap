package network

import (
	"fmt"
	"strings"
	"sync"

	"github.com/bettercap/bettercap/core"
)

type ArpTable map[string]string

var (
	arpWasParsed = false
	arpLock      = &sync.RWMutex{}
	arpTable     = make(ArpTable)
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
			ipIndex := ArpTableTokenIndex[0]
			hwIndex := ArpTableTokenIndex[1]
			ifIndex := ArpTableTokenIndex[2]

			address := m[ipIndex]
			mac := m[hwIndex]
			ifname := iface

			if ifIndex != -1 {
				ifname = m[ifIndex]
			}

			if ifname == iface {
				newTable[address] = mac
			}
		}
	}

	arpTable = newTable

	return arpTable, nil
}

func ArpLookup(iface string, address string, refresh bool) (string, error) {
	// Refresh ARP table if first run or if a force refresh has been instructed.
	if !ArpParsed() || refresh {
		if _, err := ArpUpdate(iface); err != nil {
			return "", err
		}
	}

	arpLock.RLock()
	defer arpLock.RUnlock()

	// Lookup the hardware address of this ip.
	if mac, found := arpTable[address]; found {
		return mac, nil
	}

	return "", fmt.Errorf("Could not find mac for %s", address)
}

func ArpInverseLookup(iface string, mac string, refresh bool) (string, error) {
	if !ArpParsed() || refresh {
		if _, err := ArpUpdate(iface); err != nil {
			return "", err
		}
	}

	arpLock.RLock()
	defer arpLock.RUnlock()

	for ip, hw := range arpTable {
		if hw == mac {
			return ip, nil
		}
	}

	return "", fmt.Errorf("Could not find IP for %s", mac)
}

func ArpParsed() bool {
	arpLock.RLock()
	defer arpLock.RUnlock()
	return arpWasParsed
}
