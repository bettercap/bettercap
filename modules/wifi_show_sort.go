package modules

import (
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"
)

type ByRSSISorter []*network.Station

func (a ByRSSISorter) Len() int      { return len(a) }
func (a ByRSSISorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByRSSISorter) Less(i, j int) bool {
	return a[i].RSSI < a[j].RSSI
}

type ByChannelSorter []*network.Station

func (a ByChannelSorter) Len() int      { return len(a) }
func (a ByChannelSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByChannelSorter) Less(i, j int) bool {
	return a[i].Frequency < a[j].Frequency
}

type ByBssidSorter []*network.Station

func (a ByBssidSorter) Len() int      { return len(a) }
func (a ByBssidSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByBssidSorter) Less(i, j int) bool {
	return a[i].BSSID() < a[j].BSSID()
}

type ByEssidSorter []*network.Station

func (a ByEssidSorter) Len() int      { return len(a) }
func (a ByEssidSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByEssidSorter) Less(i, j int) bool {
	if a[i].ESSID() == a[j].ESSID() {
		return a[i].HwAddress < a[j].HwAddress
	}
	return a[i].ESSID() < a[j].ESSID()
}

type ByWiFiSeenSorter []*network.Station

func (a ByWiFiSeenSorter) Len() int      { return len(a) }
func (a ByWiFiSeenSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByWiFiSeenSorter) Less(i, j int) bool {
	return a[i].LastSeen.Before(a[j].LastSeen)
}

type ByWiFiSentSorter []*network.Station

func (a ByWiFiSentSorter) Len() int      { return len(a) }
func (a ByWiFiSentSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByWiFiSentSorter) Less(i, j int) bool {
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

type ByWiFiRcvdSorter []*network.Station

func (a ByWiFiRcvdSorter) Len() int      { return len(a) }
func (a ByWiFiRcvdSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByWiFiRcvdSorter) Less(i, j int) bool {
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
