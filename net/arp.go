package net

import (
	"fmt"
	"sync"
)

type ArpTable map[string]string

var (
	arp_parsed = false
	arp_lock   = &sync.Mutex{}
	arp_table  = make(ArpTable)
)

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
