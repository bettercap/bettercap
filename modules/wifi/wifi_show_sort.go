package wifi

import (
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"
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
	return a[i].Frequency < a[j].Frequency
}

type ByEncryptionSorter []*network.Station

func (a ByEncryptionSorter) Len() int      { return len(a) }
func (a ByEncryptionSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByEncryptionSorter) Less(i, j int) bool {
	if a[i].Encryption == a[j].Encryption {
		return a[i].HwAddress < a[j].HwAddress
	}
	return a[i].Encryption < a[j].Encryption
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
	return a[i].Sent < a[j].Sent
}

type ByWiFiRcvdSorter []*network.Station

func (a ByWiFiRcvdSorter) Len() int      { return len(a) }
func (a ByWiFiRcvdSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByWiFiRcvdSorter) Less(i, j int) bool {
	return a[i].Received < a[j].Received
}

type ByClientsSorter []*network.Station

func (a ByClientsSorter) Len() int      { return len(a) }
func (a ByClientsSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByClientsSorter) Less(i, j int) bool {
	left := 0
	right := 0

	if ap, found := session.I.WiFi.Get(a[i].HwAddress); found {
		left = ap.NumClients()
	}
	if ap, found := session.I.WiFi.Get(a[j].HwAddress); found {
		right = ap.NumClients()
	}

	if left == right {
		return a[i].HwAddress < a[j].HwAddress
	}
	return left < right
}
