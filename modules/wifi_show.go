package modules

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bettercap/bettercap/network"

	"github.com/dustin/go-humanize"

	"github.com/evilsocket/islazy/ops"
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

	ssid := ops.Ternary(station.ESSID() == "<hidden>", tui.Dim(station.ESSID()), station.ESSID()).(string)

	encryption := station.Encryption
	if len(station.Cipher) > 0 {
		encryption = fmt.Sprintf("%s (%s, %s)", station.Encryption, station.Cipher, station.Authentication)
	}

	if encryption == "OPEN" || encryption == "" {
		encryption = tui.Green("OPEN")
		ssid = tui.Green(ssid)
		bssid = tui.Green(bssid)
	} else {
		// this is ugly, but necessary in order to have this
		// method handle both access point and clients
		// transparently
		if ap, found := w.Session.WiFi.Get(station.HwAddress); found && (ap.HasHandshakes() || ap.HasPMKID()) {
			encryption = tui.Red(encryption)
		}
	}

	sent := ops.Ternary(station.Sent > 0, humanize.Bytes(station.Sent), "").(string)
	recvd := ops.Ternary(station.Received > 0, humanize.Bytes(station.Received), "").(string)

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

		wps := ""
		if station.HasWPS() {
			if ver, found := station.WPS["Version"]; found {
				wps = ver
			} else {
				wps = "✔"
			}

			if state, found := station.WPS["State"]; found {
				if state == "Not Configured" {
					wps += " (not configured)"
				}
			}

			wps = tui.Dim(tui.Yellow(wps))
		}

		return []string{
			fmt.Sprintf("%d dBm", station.RSSI),
			bssid,
			ssid,
			encryption,
			wps,
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

	switch w.selector.SortField {
	case "seen":
		sort.Sort(ByWiFiSeenSorter(stations))
	case "essid":
		sort.Sort(ByEssidSorter(stations))
	case "bssid":
		sort.Sort(ByBssidSorter(stations))
	case "channel":
		sort.Sort(ByChannelSorter(stations))
	case "clients":
		sort.Sort(ByClientsSorter(stations))
	case "encryption":
		sort.Sort(ByEncryptionSorter(stations))
	case "sent":
		sort.Sort(ByWiFiSentSorter(stations))
	case "rcvd":
		sort.Sort(ByWiFiRcvdSorter(stations))
	case "rssi":
		sort.Sort(ByRSSISorter(stations))
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

func (w *WiFiModule) colDecorate(colNames []string, name string, dir string) {
	for i, c := range colNames {
		if c == name {
			colNames[i] += " " + dir
			break
		}
	}
}

func (w *WiFiModule) colNames(nrows int) []string {
	columns := []string(nil)

	if !w.isApSelected() {
		columns = []string{"RSSI", "BSSID", "SSID", "Encryption", "WPS", "Ch", "Clients", "Sent", "Recvd", "Seen"}
	} else if nrows > 0 {
		columns = []string{"RSSI", "BSSID", "Ch", "Sent", "Recvd", "Seen"}
		fmt.Printf("\n%s clients:\n", w.ap.HwAddress)
	} else {
		fmt.Printf("\nNo authenticated clients detected for %s.\n", w.ap.HwAddress)
	}

	if columns != nil {
		switch w.selector.SortField {
		case "seen":
			w.colDecorate(columns, "Seen", w.selector.SortSymbol)
		case "essid":
			w.colDecorate(columns, "SSID", w.selector.SortSymbol)
		case "bssid":
			w.colDecorate(columns, "BSSID", w.selector.SortSymbol)
		case "channel":
			w.colDecorate(columns, "Ch", w.selector.SortSymbol)
		case "clients":
			w.colDecorate(columns, "Clients", w.selector.SortSymbol)
		case "encryption":
			w.colDecorate(columns, "Encryption", w.selector.SortSymbol)
		case "sent":
			w.colDecorate(columns, "Sent", w.selector.SortSymbol)
		case "rcvd":
			w.colDecorate(columns, "Recvd", w.selector.SortSymbol)
		case "rssi":
			w.colDecorate(columns, "RSSI", w.selector.SortSymbol)
		}
	}

	return columns
}

func (w *WiFiModule) showStatusBar() {
	w.Session.Queue.Stats.RLock()
	defer w.Session.Queue.Stats.RUnlock()

	parts := []string{
		fmt.Sprintf("%s (ch. %d)", w.Session.Interface.Name(), network.GetInterfaceChannel(w.Session.Interface.Name())),
		fmt.Sprintf("%s %s", tui.Red("↑"), humanize.Bytes(w.Session.Queue.Stats.Sent)),
		fmt.Sprintf("%s %s", tui.Green("↓"), humanize.Bytes(w.Session.Queue.Stats.Received)),
		fmt.Sprintf("%d pkts", w.Session.Queue.Stats.PktReceived),
	}

	if nErrors := w.Session.Queue.Stats.Errors; nErrors > 0 {
		parts = append(parts, fmt.Sprintf("%d errs", nErrors))
	}

	if nHandshakes := w.Session.WiFi.NumHandshakes(); nHandshakes > 0 {
		parts = append(parts, fmt.Sprintf("%d handshakes", nHandshakes))
	}

	fmt.Printf("\n%s\n\n", strings.Join(parts, " / "))
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
	if nrows > 0 {
		tui.Table(os.Stdout, w.colNames(nrows), rows)
	}

	w.showStatusBar()

	w.Session.Refresh()

	return nil
}

func (w *WiFiModule) ShowWPS(bssid string) (err error) {
	toShow := []*network.Station{}

	if bssid == network.BroadcastMac {
		for _, station := range w.Session.WiFi.List() {
			if station.HasWPS() {
				toShow = append(toShow, station.Station)
			}
		}
	} else {
		if station, found := w.Session.WiFi.Get(bssid); found {
			if station.HasWPS() {
				toShow = append(toShow, station.Station)
			}
		}
	}

	if len(toShow) == 0 {
		return fmt.Errorf("no WPS enabled access points matched the criteria")
	}

	sort.Sort(ByBssidSorter(toShow))

	colNames := []string{"Name", "Value"}

	for _, station := range toShow {
		ssid := ops.Ternary(station.ESSID() == "<hidden>", tui.Dim(station.ESSID()), station.ESSID()).(string)

		rows := [][]string{
			[]string{tui.Green("essid"), ssid},
			[]string{tui.Green("bssid"), station.BSSID()},
		}

		keys := []string{}
		for name := range station.WPS {
			keys = append(keys, name)
		}
		sort.Strings(keys)

		for _, name := range keys {
			rows = append(rows, []string{
				tui.Green(name),
				tui.Yellow(station.WPS[name]),
			})
		}

		tui.Table(os.Stdout, colNames, rows)
	}

	return nil
}
