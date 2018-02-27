package modules

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"github.com/dustin/go-humanize"
)

var maxStationTTL = 5 * time.Minute

type WiFiProbe struct {
	FromAddr   net.HardwareAddr
	FromVendor string
	FromAlias  string
	SSID       string
	RSSI       int8
}

type WiFiRecon struct {
	session.SessionModule

	handle        *pcap.Handle
	channel       int
	hopPeriod     time.Duration
	frequencies   []int
	ap            *network.AccessPoint
	stickChan     int
	skipBroken    bool
	pktSourceChan chan gopacket.Packet
	writes        *sync.WaitGroup
	reads         *sync.WaitGroup
}

func NewWiFiRecon(s *session.Session) *WiFiRecon {
	w := &WiFiRecon{
		SessionModule: session.NewSessionModule("wifi.recon", s),
		channel:       0,
		stickChan:     0,
		hopPeriod:     250 * time.Millisecond,
		ap:            nil,
		skipBroken:    true,
		writes:        &sync.WaitGroup{},
		reads:         &sync.WaitGroup{},
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
			bssid, err := net.ParseMAC(args[0])
			if err != nil {
				return err
			} else if ap, found := w.Session.WiFi.Get(bssid.String()); found == true {
				w.ap = ap
				w.stickChan = mhz2chan(ap.Frequency)
				return nil
			}
			return fmt.Errorf("Could not find station with BSSID %s", args[0])
		}))

	w.AddHandler(session.NewModuleHandler("wifi.recon clear", "",
		"Remove the 802.11 base station filter.",
		func(args []string) error {
			w.ap = nil
			w.stickChan = 0
			return nil
		}))

	w.AddHandler(session.NewModuleHandler("wifi.deauth BSSID", `wifi\.deauth ((?:[0-9A-Fa-f]{2}[:-]){5}(?:[0-9A-Fa-f]{2}))`,
		"Start a 802.11 deauth attack, if an access point BSSID is provided, every client will be deauthenticated, otherwise only the selected client.",
		func(args []string) error {
			bssid, err := net.ParseMAC(args[0])
			if err != nil {
				return err
			}
			return w.startDeauth(bssid)
		}))

	w.AddHandler(session.NewModuleHandler("wifi.show", "",
		"Show current wireless stations list (default sorting by essid).",
		func(args []string) error {
			return w.Show("rssi")
		}))

	w.AddParam(session.NewIntParameter("wifi.recon.channel",
		"",
		"WiFi channel or empty for channel hopping."))

	w.AddParam(session.NewIntParameter("wifi.hop.period",
		"250",
		"If channel hopping is enabled (empty wifi.recon.channel), this is the time in millseconds the algorithm will hop on every channel (it'll be doubled if both 2.4 and 5.0 bands are available)."))

	w.AddParam(session.NewBoolParameter("wifi.skip-broken",
		"true",
		"If true, dot11 packets with an invalid checksum will be skipped."))

	return w
}

func (w WiFiRecon) Name() string {
	return "wifi.recon"
}

func (w WiFiRecon) Description() string {
	return "A module to monitor and perform wireless attacks on 802.11."
}

func (w WiFiRecon) Author() string {
	return "Gianluca Braga <matrix86@protonmail.com> && Simone Margaritelli <evilsocket@protonmail.com>>"
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
	if ssid == "<hidden>" {
		ssid = core.Dim(ssid)
	}

	encryption := station.Encryption
	if len(station.Cipher) > 0 {
		encryption = fmt.Sprintf("%s [%s,%s]", station.Encryption, station.Cipher, station.Authentication)
	}
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
			/* station.Vendor, */
			strconv.Itoa(mhz2chan(station.Frequency)),
			sent,
			recvd,
			seen,
		}
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
			strconv.Itoa(mhz2chan(station.Frequency)),
			clients,
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

func (w *WiFiRecon) isApSelected() bool {
	return w.ap != nil
}

func (w *WiFiRecon) Show(by string) error {
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
		rows = append(rows, w.getRow(s))
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

func (w *WiFiRecon) Configure() error {
	var hopPeriod int

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
	}

	if err, w.skipBroken = w.BoolParam("wifi.skip-broken"); err != nil {
		return err
	} else if err, hopPeriod = w.IntParam("wifi.hop.period"); err != nil {
		return err
	}

	w.hopPeriod = time.Duration(hopPeriod) * time.Millisecond

	if err, w.channel = w.IntParam("wifi.recon.channel"); err == nil {
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

func (w *WiFiRecon) injectPacket(data []byte) {
	if err := w.handle.WritePacketData(data); err != nil {
		log.Error("Could not send deauth packet: %s", err)

		w.Session.Queue.Stats.Lock()
		w.Session.Queue.Stats.Errors++
		w.Session.Queue.Stats.Unlock()
	} else {
		w.Session.Queue.Stats.Lock()
		w.Session.Queue.Stats.Sent += uint64(len(data))
		w.Session.Queue.Stats.Unlock()
	}
	// let the network card breath a little
	time.Sleep(10 * time.Millisecond)
}

func (w *WiFiRecon) sendDeauthPacket(ap net.HardwareAddr, client net.HardwareAddr) {
	for seq := uint16(0); seq < 64 && w.Running(); seq++ {
		if err, pkt := packets.NewDot11Deauth(ap, client, ap, seq); err != nil {
			log.Error("Could not create deauth packet: %s", err)
			continue
		} else {
			w.injectPacket(pkt)
		}

		if err, pkt := packets.NewDot11Deauth(client, ap, ap, seq); err != nil {
			log.Error("Could not create deauth packet: %s", err)
			continue
		} else {
			w.injectPacket(pkt)
		}
	}
}

func (w *WiFiRecon) onChannel(channel int, cb func()) {
	prev := w.stickChan
	w.stickChan = channel

	if err := network.SetInterfaceChannel(w.Session.Interface.Name(), channel); err != nil {
		log.Warning("Error while hopping to channel %d: %s", channel, err)
	} else {
		log.Debug("Hopped on channel %d", channel)
	}

	cb()

	w.stickChan = prev
}

func (w *WiFiRecon) startDeauth(to net.HardwareAddr) error {
	// if not already running, temporarily enable the pcap handle
	// for packet injection
	if w.Running() == false {
		if err := w.Configure(); err != nil {
			return err
		}
		defer w.handle.Close()
	}

	w.writes.Add(1)
	defer w.writes.Done()

	bssid := to.String()

	// are we deauthing every client of a given access point?
	if ap, found := w.Session.WiFi.Get(bssid); found == true {
		clients := ap.Clients()
		log.Info("Deauthing %d clients from AP %s ...", len(clients), ap.ESSID())
		w.onChannel(mhz2chan(ap.Frequency), func() {
			for _, c := range clients {
				if w.Running() == false {
					break
				}
				w.sendDeauthPacket(ap.HW, c.HW)
			}
		})

		return nil
	}

	// search for a client
	aps := w.Session.WiFi.List()
	for _, ap := range aps {
		if w.Running() == false {
			break
		} else if c, found := ap.Get(bssid); found == true {
			log.Info("Deauthing client %s from AP %s ...", c.HwAddress, ap.ESSID())
			w.onChannel(mhz2chan(ap.Frequency), func() {
				w.sendDeauthPacket(ap.HW, c.HW)
			})
			return nil
		}
	}

	return fmt.Errorf("%s is an unknown BSSID.", bssid)
}

func isZeroBSSID(bssid net.HardwareAddr) bool {
	for _, b := range bssid {
		if b != 0x00 {
			return false
		}
	}
	return true
}

func isBroadcastBSSID(bssid net.HardwareAddr) bool {
	for _, b := range bssid {
		if b != 0xff {
			return false
		}
	}
	return true
}

func (w *WiFiRecon) discoverAccessPoints(radiotap *layers.RadioTap, dot11 *layers.Dot11, packet gopacket.Packet) {
	// search for Dot11InformationElementIDSSID
	if ok, ssid := packets.Dot11ParseIDSSID(packet); ok == true {
		if isZeroBSSID(dot11.Address3) == false && isBroadcastBSSID(dot11.Address3) == false {
			bssid := dot11.Address3.String()
			frequency := int(radiotap.ChannelFrequency)
			w.Session.WiFi.AddIfNew(ssid, bssid, frequency, radiotap.DBMAntennaSignal)
		}
	}
}

func (w *WiFiRecon) discoverProbes(radiotap *layers.RadioTap, dot11 *layers.Dot11, packet gopacket.Packet) {
	if dot11.Type != layers.Dot11TypeMgmtProbeReq {
		return
	}

	reqLayer := packet.Layer(layers.LayerTypeDot11MgmtProbeReq)
	if reqLayer == nil {
		return
	}

	req, ok := reqLayer.(*layers.Dot11MgmtProbeReq)
	if ok == false {
		return
	}

	tot := len(req.Contents)
	if tot < 3 {
		return
	}

	avail := uint32(tot - 2)
	if avail < 2 {
		return
	}
	size := uint32(req.Contents[1])
	if size == 0 || size > avail {
		return
	}

	w.Session.Events.Add("wifi.client.probe", WiFiProbe{
		FromAddr:   dot11.Address2,
		FromVendor: network.OuiLookup(dot11.Address2.String()),
		FromAlias:  w.Session.Lan.GetAlias(dot11.Address2.String()),
		SSID:       string(req.Contents[2 : 2+size]),
		RSSI:       radiotap.DBMAntennaSignal,
	})
}

func (w *WiFiRecon) discoverClients(radiotap *layers.RadioTap, dot11 *layers.Dot11, packet gopacket.Packet) {
	w.Session.WiFi.EachAccessPoint(func(bssid string, ap *network.AccessPoint) {
		// packet going to this specific BSSID?
		if packets.Dot11IsDataFor(dot11, ap.HW) == true {
			ap.AddClient(dot11.Address2.String(), int(radiotap.ChannelFrequency), radiotap.DBMAntennaSignal)
		}
	})
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

	if ok, enc, cipher, auth := packets.Dot11ParseEncryption(packet, dot11); ok == true {
		bssid := dot11.Address3.String()
		if station, found := w.Session.WiFi.Get(bssid); found == true {
			station.Encryption = enc
			station.Cipher = cipher
			station.Authentication = auth
		}
	}
}

func (w *WiFiRecon) channelHopper() {
	w.reads.Add(1)
	defer w.reads.Done()

	log.Info("Channel hopper started.")
	for w.Running() == true {
		delay := w.hopPeriod
		// if we have both 2.4 and 5ghz capabilities, we have
		// more channels, therefore we need to increase the time
		// we hop on each one otherwise me lose information
		if len(w.frequencies) > 14 {
			delay = 500 * time.Millisecond
		}

		for _, frequency := range w.frequencies {
			channel := mhz2chan(frequency)
			// stick to the access point channel as long as it's selected
			// or as long as we're deauthing on it
			if w.stickChan != 0 {
				channel = w.stickChan
			}

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
	w.reads.Add(1)
	defer w.reads.Done()

	log.Debug("WiFi stations pruner started.")
	for w.Running() == true {
		for _, s := range w.Session.WiFi.List() {
			sinceLastSeen := time.Since(s.LastSeen)
			if sinceLastSeen > maxStationTTL {
				log.Debug("Station %s not seen in %s, removing.", s.BSSID(), sinceLastSeen)
				w.Session.WiFi.Remove(s.BSSID())
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (w *WiFiRecon) trackPacket(pkt gopacket.Packet) {
	pktSize := uint64(len(pkt.Data()))

	w.Session.Queue.Stats.Lock()

	w.Session.Queue.Stats.PktReceived++
	w.Session.Queue.Stats.Received += pktSize

	w.Session.Queue.Stats.Unlock()
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

		w.reads.Add(1)
		defer w.reads.Done()

		src := gopacket.NewPacketSource(w.handle, w.handle.LinkType())
		w.pktSourceChan = src.Packets()
		for packet := range w.pktSourceChan {
			if w.Running() == false {
				break
			}

			if packet == nil {
				continue
			}

			w.trackPacket(packet)

			// perform initial dot11 parsing and layers validation
			if ok, radiotap, dot11 := packets.Dot11Parse(packet); ok == true {
				// check FCS checksum
				if w.skipBroken && dot11.ChecksumValid() == false {
					log.Debug("Skipping dot11 packet with invalid checksum.")
					continue
				}

				w.discoverProbes(radiotap, dot11, packet)
				w.discoverAccessPoints(radiotap, dot11, packet)
				w.updateStats(dot11, packet)
				w.discoverClients(radiotap, dot11, packet)
			}
		}
	})

	return nil
}

func (w *WiFiRecon) Stop() error {
	return w.SetRunning(false, func() {
		// wait any pending write operation
		w.writes.Wait()
		// signal the main for loop we want to exit
		w.pktSourceChan <- nil
		// close the pcap handle to make the main for exit
		w.handle.Close()
		// close the pcap handle to make the main for exit
		// wait for the loop to exit.
		w.reads.Wait()
	})
}
