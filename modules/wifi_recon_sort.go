package modules

import (
	"github.com/bettercap/bettercap/network"
)

type ByRSSISorter []*network.Station

func (a ByRSSISorter) Len() int      { return len(a) }
func (a ByRSSISorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByRSSISorter) Less(i, j int) bool {
	if a[i].RSSI == a[j].RSSI {
		return a[i].HwAddress < a[j].HwAddress
	}
	return a[i].RSSI > a[j].RSSI
}

type ByChannelSorter []*network.Station

func (a ByChannelSorter) Len() int      { return len(a) }
func (a ByChannelSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByChannelSorter) Less(i, j int) bool {
	if a[i].Frequency == a[j].Frequency {
		return a[i].HwAddress < a[j].HwAddress
	}
	return a[i].Frequency < a[j].Frequency
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
	return a[i].LastSeen.After(a[j].LastSeen)
}
