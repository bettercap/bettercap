package net_recon

import (
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"
)

type ByAddressSorter []*network.Endpoint

func (a ByAddressSorter) Len() int      { return len(a) }
func (a ByAddressSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByAddressSorter) Less(i, j int) bool {
	if a[i].IpAddressUint32 == a[j].IpAddressUint32 {
		return a[i].HwAddress < a[j].HwAddress
	}
	return a[i].IpAddressUint32 < a[j].IpAddressUint32
}

type ByIpSorter []*network.Endpoint

func (a ByIpSorter) Len() int      { return len(a) }
func (a ByIpSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByIpSorter) Less(i, j int) bool {
	return a[i].IpAddressUint32 < a[j].IpAddressUint32
}

type ByMacSorter []*network.Endpoint

func (a ByMacSorter) Len() int      { return len(a) }
func (a ByMacSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByMacSorter) Less(i, j int) bool {
	return a[i].HwAddress < a[j].HwAddress
}

type BySeenSorter []*network.Endpoint

func (a BySeenSorter) Len() int           { return len(a) }
func (a BySeenSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySeenSorter) Less(i, j int) bool { return a[i].LastSeen.Before(a[j].LastSeen) }

type BySentSorter []*network.Endpoint

func (a BySentSorter) Len() int      { return len(a) }
func (a BySentSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a BySentSorter) Less(i, j int) bool {
	session.I.Queue.Lock()
	defer session.I.Queue.Unlock()

	var found bool = false
	var aTraffic *packets.Traffic = nil
	var bTraffic *packets.Traffic = nil

	if aTraffic, found = session.I.Queue.Traffic[a[i].IpAddress]; !found {
		aTraffic = &packets.Traffic{}
	}

	if bTraffic, found = session.I.Queue.Traffic[a[j].IpAddress]; !found {
		bTraffic = &packets.Traffic{}
	}

	return bTraffic.Sent > aTraffic.Sent
}

type ByRcvdSorter []*network.Endpoint

func (a ByRcvdSorter) Len() int      { return len(a) }
func (a ByRcvdSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByRcvdSorter) Less(i, j int) bool {
	session.I.Queue.Lock()
	defer session.I.Queue.Unlock()

	var found bool = false
	var aTraffic *packets.Traffic = nil
	var bTraffic *packets.Traffic = nil

	if aTraffic, found = session.I.Queue.Traffic[a[i].IpAddress]; !found {
		aTraffic = &packets.Traffic{}
	}

	if bTraffic, found = session.I.Queue.Traffic[a[j].IpAddress]; !found {
		bTraffic = &packets.Traffic{}
	}

	return bTraffic.Received > aTraffic.Received
}
