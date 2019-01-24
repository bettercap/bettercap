package modules

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/bettercap/bettercap/network"

	"github.com/dustin/go-humanize"

	"github.com/evilsocket/islazy/tui"
)

func (w *WiFiModule) isApSelected() bool {
	return w.ap != nil
}

func (w *WiFiModule) getRow(station *network.Station) ([]string, bool) {
	include := false
	sinceStarted := time.Since(w.Session.StartedAt)
	sinceFirstSeen := time.Since(station.FirstSeen)

	bssid := station.HwAddress
	if sinceStarted > (justJoinedTimeInterval*2) && sinceFirstSeen <= justJoinedTimeInterval {
		// if endpoint was first seen in the last 10 seconds
		bssid = tui.Bold(bssid)
	}

	seen := station.LastSeen.Format("15:04:05")
	sinceLastSeen := time.Since(station.LastSeen)
	if sinceStarted > aliveTimeInterval && sinceLastSeen <= aliveTimeInterval {
		// if endpoint seen in the last 10 seconds
		seen = tui.Bold(seen)
	} else if sinceLastSeen > presentTimeInterval {
		// if endpoint not  seen in the last 60 seconds
		seen = tui.Dim(seen)
	}

	ssid := station.ESSID()
	if ssid == "<hidden>" {
		ssid = tui.Dim(ssid)
	}

	encryption := station.Encryption
	if len(station.Cipher) > 0 {
		encryption = fmt.Sprintf("%s (%s, %s)", station.Encryption, station.Cipher, station.Authentication)
	}
	if encryption == "OPEN" || encryption == "" {
		encryption = tui.Green("OPEN")
		ssid = tui.Green(ssid)
		bssid = tui.Green(bssid)
	}
	sent := ""
	if station.Sent > 0 {
		sent = humanize.Bytes(station.Sent)
	}

	recvd := ""
	if station.Received > 0 {
		recvd = humanize.Bytes(station.Received)
	}

	if w.source == "" {
		for _, frequencies := range w.frequencies {
			if frequencies == station.Frequency {
				include = true
				break
			}
		}
	} else {
		include = true
	}

	if w.isApSelected() {
		return []string{
			fmt.Sprintf("%d dBm", station.RSSI),
			bssid,
			/* station.Vendor, */
			strconv.Itoa(station.Channel()),
			sent,
			recvd,
			seen,
		}, include
	} else {
		// this is ugly, but necessary in order to have this
		// method handle both access point and clients
		// transparently
		clients := ""
		if ap, found := w.Session.WiFi.Get(station.HwAddress); found {
			if ap.NumClients() > 0 {
				clients = strconv.Itoa(ap.NumClients())
			}
		}

		return []string{
			fmt.Sprintf("%d dBm", station.RSSI),
			bssid,
			ssid,
			/* station.Vendor, */
			encryption,
			strconv.Itoa(station.Channel()),
			clients,
			sent,
			recvd,
			seen,
		}, include
	}
}

func (w *WiFiModule) doFilter(station *network.Station) bool {
	if w.selector.Expression == nil {
		return true
	}
	return w.selector.Expression.MatchString(station.BSSID()) ||
		w.selector.Expression.MatchString(station.ESSID()) ||
		w.selector.Expression.MatchString(station.Alias) ||
		w.selector.Expression.MatchString(station.Vendor) ||
		w.selector.Expression.MatchString(station.Encryption)
}

func (w *WiFiModule) doSelection() (err error, stations []*network.Station) {
	if err = w.selector.Update(); err != nil {
		return
	}

	apSelected := w.isApSelected()
	if apSelected {
		if ap, found := w.Session.WiFi.Get(w.ap.HwAddress); found {
			stations = ap.Clients()
		} else {
			err = fmt.Errorf("Could not find station %s", w.ap.HwAddress)
			return
		}
	} else {
		stations = w.Session.WiFi.Stations()
	}

	filtered := []*network.Station{}
	for _, station := range stations {
		if w.doFilter(station) {
			filtered = append(filtered, station)
		}
	}
	stations = filtered

	// "encryption"}, "rssi"
	switch w.selector.SortBy {
	case "seen":
		sort.Sort(ByWiFiSeenSorter(stations))
	case "essid":
		sort.Sort(ByEssidSorter(stations))
	case "bssid":
		sort.Sort(ByBssidSorter(stations))
	case "channel":
		sort.Sort(ByChannelSorter(stations))
	case "sent":
		sort.Sort(ByWiFiSentSorter(stations))
	case "rcvd":
		sort.Sort(ByWiFiRcvdSorter(stations))
	case "rssi":
	default:
		sort.Sort(ByRSSISorter(stations))
	}

	// default is asc
	if w.selector.Sort == "desc" {
		// from https://github.com/golang/go/wiki/SliceTricks
		for i := len(stations)/2 - 1; i >= 0; i-- {
			opp := len(stations) - 1 - i
			stations[i], stations[opp] = stations[opp], stations[i]
		}
	}

	if w.selector.Limit > 0 {
		limit := w.selector.Limit
		max := len(stations)
		if limit > max {
			limit = max
		}
		stations = stations[0:limit]
	}

	return
}

func (w *WiFiModule) Show() (err error) {
	var stations []*network.Station
	if err, stations = w.doSelection(); err != nil {
		return
	}

	rows := make([][]string, 0)
	for _, s := range stations {
		if row, include := w.getRow(s); include {
			rows = append(rows, row)
		}
	}
	nrows := len(rows)

	columns := []string{"RSSI", "BSSID", "SSID" /* "Vendor", */, "Encryption", "Channel", "Clients", "Sent", "Recvd", "Last Seen"}
	if w.isApSelected() {
		// these are clients
		columns = []string{"RSSI", "MAC" /* "Vendor", */, "Channel", "Sent", "Received", "Last Seen"}

		if nrows == 0 {
			fmt.Printf("\nNo authenticated clients detected for %s.\n", w.ap.HwAddress)
		} else {
			fmt.Printf("\n%s clients:\n", w.ap.HwAddress)
		}
	}

	if nrows > 0 {
		tui.Table(os.Stdout, columns, rows)
	}

	w.Session.Queue.Stats.RLock()
	fmt.Printf("\n%s (ch. %d) / %s %s / %s %s / %d pkts / %d errs\n\n",
		w.Session.Interface.Name(),
		network.GetInterfaceChannel(w.Session.Interface.Name()),
		tui.Red("↑"),
		humanize.Bytes(w.Session.Queue.Stats.Sent),
		tui.Green("↓"),
		humanize.Bytes(w.Session.Queue.Stats.Received),
		w.Session.Queue.Stats.PktReceived,
		w.Session.Queue.Stats.Errors)
	w.Session.Queue.Stats.RUnlock()

	w.Session.Refresh()

	return nil
}
