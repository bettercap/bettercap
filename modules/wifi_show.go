package modules

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/network"

	"github.com/dustin/go-humanize"
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
		bssid = core.Bold(bssid)
	}

	seen := station.LastSeen.Format("15:04:05")
	sinceLastSeen := time.Since(station.LastSeen)
	if sinceStarted > aliveTimeInterval && sinceLastSeen <= aliveTimeInterval {
		// if endpoint seen in the last 10 seconds
		seen = core.Bold(seen)
	} else if sinceLastSeen > presentTimeInterval {
		// if endpoint not  seen in the last 60 seconds
		seen = core.Dim(seen)
	}

	ssid := station.ESSID()
	if ssid == "<hidden>" {
		ssid = core.Dim(ssid)
	}

	encryption := station.Encryption
	if len(station.Cipher) > 0 {
		encryption = fmt.Sprintf("%s (%s, %s)", station.Encryption, station.Cipher, station.Authentication)
	}
	if encryption == "OPEN" || encryption == "" {
		encryption = core.Green("OPEN")
		ssid = core.Green(ssid)
		bssid = core.Green(bssid)
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
			strconv.Itoa(network.Dot11Freq2Chan(station.Frequency)),
			sent,
			recvd,
			seen,
		}, include
	} else {
		// this is ugly, but necessary in order to have this
		// method handle both access point and clients
		// transparently
		clients := ""
		if ap, found := w.Session.WiFi.Get(station.HwAddress); found == true {
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
			strconv.Itoa(network.Dot11Freq2Chan(station.Frequency)),
			clients,
			sent,
			recvd,
			seen,
		}, include
	}
}

func (w *WiFiModule) Show(by string) error {
	var stations []*network.Station

	apSelected := w.isApSelected()
	if apSelected {
		if ap, found := w.Session.WiFi.Get(w.ap.HwAddress); found == true {
			stations = ap.Clients()
		} else {
			return fmt.Errorf("Could not find station %s", w.ap.HwAddress)
		}
	} else {
		stations = w.Session.WiFi.Stations()
	}

	if by == "seen" {
		sort.Sort(ByWiFiSeenSorter(stations))
	} else if by == "essid" {
		sort.Sort(ByEssidSorter(stations))
	} else if by == "channel" {
		sort.Sort(ByChannelSorter(stations))
	} else {
		sort.Sort(ByRSSISorter(stations))
	}

	rows := make([][]string, 0)
	for _, s := range stations {
		if row, include := w.getRow(s); include == true {
			rows = append(rows, row)
		}
	}
	nrows := len(rows)

	columns := []string{"RSSI", "BSSID", "SSID" /* "Vendor", */, "Encryption", "Channel", "Clients", "Sent", "Recvd", "Last Seen"}
	if apSelected {
		// these are clients
		columns = []string{"RSSI", "MAC" /* "Vendor", */, "Channel", "Sent", "Received", "Last Seen"}

		if nrows == 0 {
			fmt.Printf("\nNo authenticated clients detected for %s.\n", w.ap.HwAddress)
		} else {
			fmt.Printf("\n%s clients:\n", w.ap.HwAddress)
		}
	}

	if nrows > 0 {
		core.AsTable(os.Stdout, columns, rows)
	}

	w.Session.Refresh()

	return nil
}
