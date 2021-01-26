package wifi

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bettercap/bettercap/modules/net_recon"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"

	"github.com/dustin/go-humanize"

	"github.com/evilsocket/islazy/ops"
	"github.com/evilsocket/islazy/tui"
)

func (mod *WiFiModule) isApSelected() bool {
	return mod.ap != nil
}

func (mod *WiFiModule) getRow(station *network.Station) ([]string, bool) {
	rssi := network.ColorRSSI(int(station.RSSI))
	bssid := station.HwAddress
	sinceStarted := time.Since(mod.Session.StartedAt)
	sinceFirstSeen := time.Since(station.FirstSeen)
	if sinceStarted > (net_recon.JustJoinedTimeInterval*2) && sinceFirstSeen <= net_recon.JustJoinedTimeInterval {
		// if endpoint was first seen in the last 10 seconds
		bssid = tui.Bold(bssid)
	}

	seen := station.LastSeen.Format("15:04:05")
	sinceLastSeen := time.Since(station.LastSeen)
	if sinceStarted > net_recon.AliveTimeInterval && sinceLastSeen <= net_recon.AliveTimeInterval {
		// if endpoint seen in the last 10 seconds
		seen = tui.Bold(seen)
	} else if sinceLastSeen > net_recon.PresentTimeInterval {
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
		if ap, found := mod.Session.WiFi.Get(station.HwAddress); found && ap.HasKeyMaterial() {
			encryption = tui.Red(encryption)
		}
	}

	sent := ops.Ternary(station.Sent > 0, humanize.Bytes(station.Sent), "").(string)
	recvd := ops.Ternary(station.Received > 0, humanize.Bytes(station.Received), "").(string)

	include := false
	if mod.source == "" {
		for _, frequencies := range mod.frequencies {
			if frequencies == station.Frequency {
				include = true
				break
			}
		}
	} else {
		include = true
	}

	if int(station.RSSI) < mod.minRSSI {
		include = false
	}

	if mod.isApSelected() {
		if mod.showManuf {
			return []string{
				rssi,
				bssid,
				tui.Dim(station.Vendor),
				strconv.Itoa(station.Channel),
				sent,
				recvd,
				seen,
			}, include
		} else {
			return []string{
				rssi,
				bssid,
				strconv.Itoa(station.Channel),
				sent,
				recvd,
				seen,
			}, include
		}
	} else {
		// this is ugly, but necessary in order to have this
		// method handle both access point and clients
		// transparently
		clients := ""
		if ap, found := mod.Session.WiFi.Get(station.HwAddress); found {
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

		if mod.showManuf {
			return []string{
				rssi,
				bssid,
				tui.Dim(station.Vendor),
				ssid,
				encryption,
				wps,
				strconv.Itoa(station.Channel),
				clients,
				sent,
				recvd,
				seen,
			}, include
		} else {
			return []string{
				rssi,
				bssid,
				ssid,
				encryption,
				wps,
				strconv.Itoa(station.Channel),
				clients,
				sent,
				recvd,
				seen,
			}, include
		}
	}
}

func (mod *WiFiModule) doFilter(station *network.Station) bool {
	if mod.selector.Expression == nil {
		return true
	}
	return mod.selector.Expression.MatchString(station.BSSID()) ||
		mod.selector.Expression.MatchString(station.ESSID()) ||
		mod.selector.Expression.MatchString(station.Alias) ||
		mod.selector.Expression.MatchString(station.Vendor) ||
		mod.selector.Expression.MatchString(station.Encryption)
}

func (mod *WiFiModule) doSelection() (err error, stations []*network.Station) {
	if err = mod.selector.Update(); err != nil {
		return
	}

	apSelected := mod.isApSelected()
	if apSelected {
		if ap, found := mod.Session.WiFi.Get(mod.ap.HwAddress); found {
			stations = ap.Clients()
		} else {
			err = fmt.Errorf("Could not find station %s", mod.ap.HwAddress)
			return
		}
	} else {
		stations = mod.Session.WiFi.Stations()
	}

	filtered := []*network.Station{}
	for _, station := range stations {
		if mod.doFilter(station) {
			filtered = append(filtered, station)
		}
	}
	stations = filtered

	switch mod.selector.SortField {
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
	if mod.selector.Sort == "desc" {
		// from https://github.com/golang/go/wiki/SliceTricks
		for i := len(stations)/2 - 1; i >= 0; i-- {
			opp := len(stations) - 1 - i
			stations[i], stations[opp] = stations[opp], stations[i]
		}
	}

	if mod.selector.Limit > 0 {
		limit := mod.selector.Limit
		max := len(stations)
		if limit > max {
			limit = max
		}
		stations = stations[0:limit]
	}

	return
}

func (mod *WiFiModule) colDecorate(colNames []string, name string, dir string) {
	for i, c := range colNames {
		if c == name {
			colNames[i] += " " + dir
			break
		}
	}
}

func (mod *WiFiModule) colNames(nrows int) []string {
	columns := []string(nil)

	if !mod.isApSelected() {
		if mod.showManuf {
			columns = []string{"RSSI", "BSSID", "Manufacturer", "SSID", "Encryption", "WPS", "Ch", "Clients", "Sent", "Recvd", "Seen"}
		} else {
			columns = []string{"RSSI", "BSSID", "SSID", "Encryption", "WPS", "Ch", "Clients", "Sent", "Recvd", "Seen"}
		}
	} else if nrows > 0 {
		if mod.showManuf {
			columns = []string{"RSSI", "BSSID", "Manufacturer", "Ch", "Sent", "Recvd", "Seen"}
		} else {
			columns = []string{"RSSI", "BSSID", "Ch", "Sent", "Recvd", "Seen"}
		}
		mod.Printf("\n%s clients:\n", mod.ap.HwAddress)
	} else {
		mod.Printf("\nNo authenticated clients detected for %s.\n", mod.ap.HwAddress)
	}

	if columns != nil {
		switch mod.selector.SortField {
		case "seen":
			mod.colDecorate(columns, "Seen", mod.selector.SortSymbol)
		case "essid":
			mod.colDecorate(columns, "SSID", mod.selector.SortSymbol)
		case "bssid":
			mod.colDecorate(columns, "BSSID", mod.selector.SortSymbol)
		case "channel":
			mod.colDecorate(columns, "Ch", mod.selector.SortSymbol)
		case "clients":
			mod.colDecorate(columns, "Clients", mod.selector.SortSymbol)
		case "encryption":
			mod.colDecorate(columns, "Encryption", mod.selector.SortSymbol)
		case "sent":
			mod.colDecorate(columns, "Sent", mod.selector.SortSymbol)
		case "rcvd":
			mod.colDecorate(columns, "Recvd", mod.selector.SortSymbol)
		case "rssi":
			mod.colDecorate(columns, "RSSI", mod.selector.SortSymbol)
		}
	}

	return columns
}

func (mod *WiFiModule) showStatusBar() {
	parts := []string{
		fmt.Sprintf("%s (ch. %d)", mod.iface.Name(), network.GetInterfaceChannel(mod.iface.Name())),
		fmt.Sprintf("%s %s", tui.Red("↑"), humanize.Bytes(mod.Session.Queue.Stats.Sent)),
		fmt.Sprintf("%s %s", tui.Green("↓"), humanize.Bytes(mod.Session.Queue.Stats.Received)),
		fmt.Sprintf("%d pkts", mod.Session.Queue.Stats.PktReceived),
	}

	if nErrors := mod.Session.Queue.Stats.Errors; nErrors > 0 {
		parts = append(parts, fmt.Sprintf("%d errs", nErrors))
	}

	if nHandshakes := mod.Session.WiFi.NumHandshakes(); nHandshakes > 0 {
		parts = append(parts, fmt.Sprintf("%d handshakes", nHandshakes))
	}

	mod.Printf("\n%s\n\n", strings.Join(parts, " / "))
}

func (mod *WiFiModule) Show() (err error) {
	if mod.Running() == false {
		return session.ErrAlreadyStopped(mod.Name())
	}

	var stations []*network.Station
	if err, stations = mod.doSelection(); err != nil {
		return
	}

	if err, mod.showManuf = mod.BoolParam("wifi.show.manufacturer"); err != nil {
		return err
	}

	rows := make([][]string, 0)
	for _, s := range stations {
		if row, include := mod.getRow(s); include {
			rows = append(rows, row)
		}
	}
	nrows := len(rows)
	if nrows > 0 {
		tui.Table(mod.Session.Events.Stdout, mod.colNames(nrows), rows)
	}

	mod.showStatusBar()

	mod.Session.Refresh()

	return nil
}

func (mod *WiFiModule) ShowWPS(bssid string) (err error) {
	if mod.Running() == false {
		return session.ErrAlreadyStopped(mod.Name())
	}

	toShow := []*network.Station{}
	if bssid == network.BroadcastMac {
		for _, station := range mod.Session.WiFi.List() {
			if station.HasWPS() {
				toShow = append(toShow, station.Station)
			}
		}
	} else {
		if station, found := mod.Session.WiFi.Get(bssid); found {
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
			{tui.Green("essid"), ssid},
			{tui.Green("bssid"), station.BSSID()},
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

		tui.Table(mod.Session.Events.Stdout, colNames, rows)
	}

	return nil
}
