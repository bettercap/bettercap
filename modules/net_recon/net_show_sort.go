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

func trafficOf(ip string) *packets.Traffic {
	if v, found := session.I.Queue.Traffic.Load(ip); !found {
		return &packets.Traffic{}
	} else {
		return v.(*packets.Traffic)
	}
}

func (a BySentSorter) Len() int      { return len(a) }
func (a BySentSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a BySentSorter) Less(i, j int) bool {
	aTraffic := trafficOf(a[i].IpAddress)
	bTraffic := trafficOf(a[j].IpAddress)
	return bTraffic.Sent > aTraffic.Sent
}

type ByRcvdSorter []*network.Endpoint

func (a ByRcvdSorter) Len() int      { return len(a) }
func (a ByRcvdSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByRcvdSorter) Less(i, j int) bool {
	aTraffic := trafficOf(a[i].IpAddress)
	bTraffic := trafficOf(a[j].IpAddress)
	return bTraffic.Received > aTraffic.Received
}
