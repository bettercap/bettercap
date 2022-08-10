package wifi

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/bettercap/bettercap/modules/utils"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/ops"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

type WiFiModule struct {
	session.SessionModule

	iface               *network.Endpoint
	handle              *pcap.Handle
	source              string
	region              string
	txPower             int
	minRSSI             int
	apTTL               int
	staTTL              int
	channel             int
	hopPeriod           time.Duration
	hopChanges          chan bool
	frequencies         []int
	ap                  *network.AccessPoint
	stickChan           int
	shakesFile          string
	shakesAggregate     bool
	skipBroken          bool
	pktSourceChan       chan gopacket.Packet
	pktSourceChanClosed bool
	deauthSkip          []net.HardwareAddr
	deauthSilent        bool
	deauthOpen          bool
	deauthAcquired      bool
	assocSkip           []net.HardwareAddr
	assocSilent         bool
	assocOpen           bool
	assocAcquired       bool
	csaSilent           bool
	fakeAuthSilent      bool
	filterProbeSTA      *regexp.Regexp
	filterProbeAP       *regexp.Regexp
	apRunning           bool
	showManuf           bool
	apConfig            packets.Dot11ApConfig
	probeMac            net.HardwareAddr
	writes              *sync.WaitGroup
	reads               *sync.WaitGroup
	chanLock            *sync.Mutex
	selector            *utils.ViewSelector
}

func NewWiFiModule(s *session.Session) *WiFiModule {
	mod := &WiFiModule{
		SessionModule:   session.NewSessionModule("wifi", s),
		iface:           s.Interface,
		minRSSI:         -200,
		apTTL:           300,
		staTTL:          300,
		channel:         0,
		stickChan:       0,
		hopPeriod:       250 * time.Millisecond,
		hopChanges:      make(chan bool),
		ap:              nil,
		skipBroken:      true,
		apRunning:       false,
		deauthSkip:      []net.HardwareAddr{},
		deauthSilent:    false,
		deauthOpen:      false,
		deauthAcquired:  false,
		assocSkip:       []net.HardwareAddr{},
		assocSilent:     false,
		assocOpen:       false,
		assocAcquired:   false,
		csaSilent:       false,
		fakeAuthSilent:  false,
		showManuf:       false,
		shakesAggregate: true,
		writes:          &sync.WaitGroup{},
		reads:           &sync.WaitGroup{},
		chanLock:        &sync.Mutex{},
	}

	mod.InitState("channels")

	mod.AddParam(session.NewStringParameter("wifi.interface",
		"",
		"",
		"If filled, will use this interface name instead of the one provided by the -iface argument or detected automatically."))

	mod.AddHandler(session.NewModuleHandler("wifi.recon on", "",
		"Start 802.11 wireless base stations discovery and channel hopping.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("wifi.recon off", "",
		"Stop 802.11 wireless base stations discovery and channel hopping.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("wifi.clear", "",
		"Clear all access points collected by the WiFi discovery module.",
		func(args []string) error {
			mod.Session.WiFi.Clear()
			return nil
		}))

	mod.AddHandler(session.NewModuleHandler("wifi.recon MAC", "wifi.recon ((?:[0-9A-Fa-f]{2}[:-]){5}(?:[0-9A-Fa-f]{2}))",
		"Set 802.11 base station address to filter for.",
		func(args []string) error {
			bssid, err := net.ParseMAC(args[0])
			if err != nil {
				return err
			} else if ap, found := mod.Session.WiFi.Get(bssid.String()); found {
				mod.ap = ap
				mod.stickChan = ap.Channel
				return nil
			}
			return fmt.Errorf("Could not find station with BSSID %s", args[0])
		}))

	mod.AddHandler(session.NewModuleHandler("wifi.recon clear", "",
		"Remove the 802.11 base station filter.",
		func(args []string) (err error) {
			mod.ap = nil
			mod.stickChan = 0
			freqs, err := network.GetSupportedFrequencies(mod.iface.Name())
			mod.setFrequencies(freqs)
			mod.hopChanges <- true
			return err
		}))

	mod.AddHandler(session.NewModuleHandler("wifi.client.probe.sta.filter FILTER", "wifi.client.probe.sta.filter (.+)",
		"Use this regular expression on the station address to filter client probes, 'clear' to reset the filter.",
		func(args []string) (err error) {
			filter := args[0]
			if filter == "clear" {
				mod.filterProbeSTA = nil
				return
			} else if mod.filterProbeSTA, err = regexp.Compile(filter); err != nil {
				return
			}
			return
		}))

	mod.AddHandler(session.NewModuleHandler("wifi.client.probe.ap.filter FILTER", "wifi.client.probe.ap.filter (.+)",
		"Use this regular expression on the access point name to filter client probes, 'clear' to reset the filter.",
		func(args []string) (err error) {
			filter := args[0]
			if filter == "clear" {
				mod.filterProbeAP = nil
				return
			} else if mod.filterProbeAP, err = regexp.Compile(filter); err != nil {
				return
			}
			return
		}))

	minRSSI := session.NewIntParameter("wifi.rssi.min",
		"-200",
		"Minimum WiFi signal strength in dBm.")

	mod.AddObservableParam(minRSSI, func(v string) {
		if err, v := minRSSI.Get(s); err != nil {
			mod.Error("%v", err)
		} else if mod.minRSSI = v.(int); mod.Started {
			mod.Info("wifi.rssi.min set to %d", mod.minRSSI)
		}
	})

	deauth := session.NewModuleHandler("wifi.deauth BSSID", `wifi\.deauth ((?:[a-fA-F0-9:]{11,})|all|\*)`,
		"Start a 802.11 deauth attack, if an access point BSSID is provided, every client will be deauthenticated, otherwise only the selected client. Use 'all', '*' or a broadcast BSSID (ff:ff:ff:ff:ff:ff) to iterate every access point with at least one client and start a deauth attack for each one.",
		func(args []string) error {
			if args[0] == "all" || args[0] == "*" {
				args[0] = "ff:ff:ff:ff:ff:ff"
			}
			bssid, err := net.ParseMAC(args[0])
			if err != nil {
				return err
			}
			return mod.startDeauth(bssid)
		})

	deauth.Complete("wifi.deauth", s.WiFiCompleterFull)

	mod.AddHandler(deauth)

	probe := session.NewModuleHandler("wifi.probe BSSID ESSID",
		`wifi\.probe\s+([a-fA-F0-9:]{11,})\s+([^\s].+)`,
		"Sends a fake client probe with the given station BSSID, searching for ESSID.",
		func(args []string) (err error) {
			if mod.probeMac, err = net.ParseMAC(args[0]); err != nil {
				return err
			}
			return mod.startProbing(mod.probeMac, args[1])
		})

	probe.Complete("wifi.probe", s.WiFiCompleterFull)

	mod.AddHandler(probe)

	channelSwitchAnnounce := session.NewModuleHandler("wifi.channel_switch_announce bssid channel ", `wifi\.channel_switch_announce ((?:[a-fA-F0-9:]{11,}))\s+((?:[0-9]+))`,
		"Start a 802.11 channel hop attack, all client will be force to change the channel lead to connection down.",
		func(args []string) error {
			bssid, err := net.ParseMAC(args[0])
			if err != nil {
				return err
			}
			channel, _ := strconv.Atoi(args[1])
			if channel > 180 || channel < 1 {
				return fmt.Errorf("%d is not a valid channel number", channel)
			}
			return mod.startCSA(bssid, int8(channel))
		})

	channelSwitchAnnounce.Complete("wifi.channel_switch_announce", s.WiFiCompleterFull)

	mod.AddHandler(channelSwitchAnnounce)

	fakeAuth := session.NewModuleHandler("wifi.fake_auth bssid client", `wifi\.fake_auth ((?:[a-fA-F0-9:]{11,}))\s+((?:[a-fA-F0-9:]{11,}))`,
		"send an fake authentication with client mac to ap lead to client disconnect",
		func(args []string) error {
			bssid, err := net.ParseMAC(args[0])
			if err != nil {
				return err
			}
			client, err := net.ParseMAC(args[1])
			if err != nil {
				return err
			}
			return mod.startFakeAuth(bssid, client)
		})

	fakeAuth.Complete("wifi.fake_auth", s.WiFiCompleterFull)

	mod.AddHandler(fakeAuth)

	mod.AddParam(session.NewBoolParameter("wifi.channel_switch_announce.silent",
		"false",
		"If true, messages from wifi.channel_switch_announce will be suppressed."))

	mod.AddParam(session.NewBoolParameter("wifi.fake_auth.silent",
		"false",
		"If true, messages from wifi.fake_auth will be suppressed."))

	mod.AddParam(session.NewStringParameter("wifi.deauth.skip",
		"",
		"",
		"Comma separated list of BSSID to skip while sending deauth packets."))

	mod.AddParam(session.NewBoolParameter("wifi.deauth.silent",
		"false",
		"If true, messages from wifi.deauth will be suppressed."))

	mod.AddParam(session.NewBoolParameter("wifi.deauth.open",
		"true",
		"Send wifi deauth packets to open networks."))

	mod.AddParam(session.NewBoolParameter("wifi.deauth.acquired",
		"false",
		"Send wifi deauth packets from AP's for which key material was already acquired."))

	assoc := session.NewModuleHandler("wifi.assoc BSSID", `wifi\.assoc ((?:[a-fA-F0-9:]{11,})|all|\*)`,
		"Send an association request to the selected BSSID in order to receive a RSN PMKID key. Use 'all', '*' or a broadcast BSSID (ff:ff:ff:ff:ff:ff) to iterate for every access point.",
		func(args []string) error {
			if args[0] == "all" || args[0] == "*" {
				args[0] = "ff:ff:ff:ff:ff:ff"
			}
			bssid, err := net.ParseMAC(args[0])
			if err != nil {
				return err
			}
			return mod.startAssoc(bssid)
		})

	assoc.Complete("wifi.assoc", s.WiFiCompleter)

	mod.AddHandler(assoc)

	apTTL := session.NewIntParameter("wifi.ap.ttl",
		"300",
		"Seconds of inactivity for an access points to be considered not in range anymore.")

	mod.AddObservableParam(apTTL, func(v string) {
		if err, v := apTTL.Get(s); err != nil {
			mod.Error("%v", err)
		} else if mod.apTTL = v.(int); mod.Started {
			mod.Info("wifi.ap.ttl set to %d", mod.apTTL)
		}
	})

	staTTL := session.NewIntParameter("wifi.sta.ttl",
		"300",
		"Seconds of inactivity for a client station to be considered not in range or not connected to its access point anymore.")

	mod.AddObservableParam(staTTL, func(v string) {
		if err, v := staTTL.Get(s); err != nil {
			mod.Error("%v", err)
		} else if mod.staTTL = v.(int); mod.Started {
			mod.Info("wifi.sta.ttl set to %d", mod.staTTL)
		}
	})

	mod.AddParam(session.NewStringParameter("wifi.region",
		"",
		"",
		"Set the WiFi region to this value before activating the interface."))

	mod.AddParam(session.NewIntParameter("wifi.txpower",
		"30",
		"Set WiFi transmission power to this value before activating the interface."))

	mod.AddParam(session.NewStringParameter("wifi.assoc.skip",
		"",
		"",
		"Comma separated list of BSSID to skip while sending association requests."))

	mod.AddParam(session.NewBoolParameter("wifi.assoc.silent",
		"false",
		"If true, messages from wifi.assoc will be suppressed."))

	mod.AddParam(session.NewBoolParameter("wifi.assoc.open",
		"false",
		"Send association requests to open networks."))

	mod.AddParam(session.NewBoolParameter("wifi.assoc.acquired",
		"false",
		"Send association to AP's for which key material was already acquired."))

	mod.AddHandler(session.NewModuleHandler("wifi.ap", "",
		"Inject fake management beacons in order to create a rogue access point.",
		func(args []string) error {
			if err := mod.parseApConfig(); err != nil {
				return err
			} else {
				return mod.startAp()
			}
		}))

	mod.AddParam(session.NewStringParameter("wifi.handshakes.file",
		"~/bettercap-wifi-handshakes.pcap",
		"",
		"File path of the pcap file to save handshakes to."))

	mod.AddParam(session.NewBoolParameter("wifi.handshakes.aggregate",
		"true",
		"If true, all handshakes will be saved inside a single file, otherwise a folder with per-network pcap files will be created."))

	mod.AddParam(session.NewStringParameter("wifi.ap.ssid",
		"FreeWiFi",
		"",
		"SSID of the fake access point."))

	mod.AddParam(session.NewStringParameter("wifi.ap.bssid",
		session.ParamRandomMAC,
		"[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}",
		"BSSID of the fake access point."))

	mod.AddParam(session.NewIntParameter("wifi.ap.channel",
		"1",
		"Channel of the fake access point."))

	mod.AddParam(session.NewBoolParameter("wifi.ap.encryption",
		"true",
		"If true, the fake access point will use WPA2, otherwise it'll result as an open AP."))

	mod.AddHandler(session.NewModuleHandler("wifi.show.wps BSSID",
		`wifi\.show\.wps ((?:[a-fA-F0-9:]{11,})|all|\*)`,
		"Show WPS information about a given station (use 'all', '*' or a broadcast BSSID for all).",
		func(args []string) error {
			if args[0] == "all" || args[0] == "*" {
				args[0] = "ff:ff:ff:ff:ff:ff"
			}
			return mod.ShowWPS(args[0])
		}))

	mod.AddHandler(session.NewModuleHandler("wifi.show", "",
		"Show current wireless stations list (default sorting by essid).",
		func(args []string) error {
			return mod.Show()
		}))

	mod.selector = utils.ViewSelectorFor(&mod.SessionModule, "wifi.show",
		[]string{"rssi", "bssid", "essid", "channel", "encryption", "clients", "seen", "sent", "rcvd"}, "rssi asc")

	mod.AddParam(session.NewBoolParameter("wifi.show.manufacturer",
		"false",
		"If true, wifi.show will also show the devices manufacturers."))

	mod.AddHandler(session.NewModuleHandler("wifi.recon.channel CHANNEL", `wifi\.recon\.channel[\s]+([0-9]+(?:[, ]+[0-9]+)*|clear)`,
		"WiFi channels (comma separated) or 'clear' for channel hopping.",
		func(args []string) (err error) {
			freqs := []int{}

			if args[0] != "clear" {
				mod.Debug("setting hopping channels to %s", args[0])
				for _, s := range str.Comma(args[0]) {
					if ch, err := strconv.Atoi(s); err != nil {
						return err
					} else {
						if f := network.Dot11Chan2Freq(ch); f == 0 {
							return fmt.Errorf("%d is not a valid wifi channel.", ch)
						} else {
							freqs = append(freqs, f)
						}
					}
				}
			}

			if len(freqs) == 0 {
				mod.Debug("resetting hopping channels")
				if mod.iface == nil {
					return fmt.Errorf("wifi.interface not set or not found")
				} else if freqs, err = network.GetSupportedFrequencies(mod.iface.Name()); err != nil {
					return err
				}
			}

			mod.setFrequencies(freqs)

			// if wifi.recon is not running, this would block forever
			if mod.Running() {
				mod.hopChanges <- true
			}

			return nil
		}))

	mod.AddParam(session.NewStringParameter("wifi.source.file",
		"",
		"",
		"If set, the wifi module will read from this pcap file instead of the hardware interface."))

	mod.AddParam(session.NewIntParameter("wifi.hop.period",
		"250",
		"If channel hopping is enabled (empty wifi.recon.channel), this is the time in milliseconds the algorithm will hop on every channel (it'll be doubled if both 2.4 and 5.0 bands are available)."))

	mod.AddParam(session.NewBoolParameter("wifi.skip-broken",
		"true",
		"If true, dot11 packets with an invalid checksum will be skipped."))

	return mod
}

func (mod WiFiModule) Name() string {
	return "wifi"
}

func (mod WiFiModule) Description() string {
	return "A module to monitor and perform wireless attacks on 802.11."
}

func (mod WiFiModule) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com> && Gianluca Braga <matrix86@gmail.com>"
}

const (
	// Ugly, but gopacket folks are not exporting pcap errors, so ...
	// ref. https://github.com/google/gopacket/blob/96986c90e3e5c7e01deed713ff8058e357c0c047/pcap/pcap.go#L281
	ErrIfaceNotUp = "Interface Not Up"
)

func (mod *WiFiModule) setFrequencies(freqs []int) {
	mod.Debug("new frequencies: %v", freqs)

	mod.frequencies = freqs
	channels := []int{}
	for _, freq := range freqs {
		channels = append(channels, network.Dot11Freq2Chan(freq))
	}
	mod.State.Store("channels", channels)
}

func (mod *WiFiModule) Configure() error {
	var ifName string
	var hopPeriod int
	var err error

	if err, mod.apTTL = mod.IntParam("wifi.ap.ttl"); err != nil {
		return err
	} else if err, mod.staTTL = mod.IntParam("wifi.sta.ttl"); err != nil {
		return err
	}

	if err, mod.region = mod.StringParam("wifi.region"); err != nil {
		return err
	} else if err, mod.txPower = mod.IntParam("wifi.txpower"); err != nil {
		return err
	} else if err, mod.source = mod.StringParam("wifi.source.file"); err != nil {
		return err
	} else if err, mod.minRSSI = mod.IntParam("wifi.rssi.min"); err != nil {
		return err
	}

	if err, mod.shakesAggregate = mod.BoolParam("wifi.handshakes.aggregate"); err != nil {
		return err
	} else if err, mod.shakesFile = mod.StringParam("wifi.handshakes.file"); err != nil {
		return err
	} else if mod.shakesFile != "" {
		if mod.shakesFile, err = fs.Expand(mod.shakesFile); err != nil {
			return err
		}
	}

	if err, ifName = mod.StringParam("wifi.interface"); err != nil {
		return err
	} else if ifName == "" {
		mod.iface = mod.Session.Interface
		ifName = mod.iface.Name()
	} else if mod.iface, err = network.FindInterface(ifName); err != nil {
		return fmt.Errorf("could not find interface %s: %v", ifName, err)
	} else if mod.iface == nil {
		return fmt.Errorf("could not find interface %s", ifName)
	}

	mod.Info("using interface %s (%s)", ifName, mod.iface.HwAddress)

	if mod.source != "" {
		if mod.handle, err = pcap.OpenOffline(mod.source); err != nil {
			return fmt.Errorf("error while opening file %s: %s", mod.source, err)
		}
	} else {
		if mod.region != "" {
			if err := network.SetWiFiRegion(mod.region); err != nil {
				return err
			} else {
				mod.Debug("WiFi region set to '%s'", mod.region)
			}
		}

		if mod.txPower > 0 {
			if err := network.SetInterfaceTxPower(ifName, mod.txPower); err != nil {
				mod.Warning("could not set interface %s txpower to %d, 'Set Tx Power' requests not supported: %v", ifName, mod.txPower, err)
			} else {
				mod.Debug("interface %s txpower set to %d", ifName, mod.txPower)
			}
		}

		/*
		 * We don't want to pcap.BlockForever otherwise pcap_close(handle)
		 * could hang waiting for a timeout to expire ...
		 */
		opts := network.CAPTURE_DEFAULTS
		opts.Timeout = 500 * time.Millisecond
		opts.Monitor = true

		for retry := 0; ; retry++ {
			if mod.handle, err = network.CaptureWithOptions(ifName, opts); err == nil {
				// we're done
				break
			} else if retry == 0 && err.Error() == ErrIfaceNotUp {
				// try to bring interface up and try again
				mod.Info("interface %s is down, bringing it up ...", ifName)
				if err := network.ActivateInterface(ifName); err != nil {
					return err
				}
				continue
			} else if !opts.Monitor {
				// second fatal error, just bail
				return fmt.Errorf("error while activating handle: %s", err)
			} else {
				// first fatal error, try again without setting the interface in monitor mode
				mod.Warning("error while activating handle: %s, %s", err, tui.Bold("interface might already be monitoring. retrying!"))
				opts.Monitor = false
			}
		}
	}

	if err, mod.skipBroken = mod.BoolParam("wifi.skip-broken"); err != nil {
		return err
	} else if err, hopPeriod = mod.IntParam("wifi.hop.period"); err != nil {
		return err
	}

	mod.hopPeriod = time.Duration(hopPeriod) * time.Millisecond

	if mod.source == "" {
		if freqs, err := network.GetSupportedFrequencies(ifName); err != nil {
			return fmt.Errorf("error while getting supported frequencies of %s: %s", ifName, err)
		} else {
			mod.setFrequencies(freqs)
		}

		mod.Debug("wifi supported frequencies: %v", mod.frequencies)

		// we need to start somewhere, this is just to check if
		// this OS supports switching channel programmatically.
		if err = network.SetInterfaceChannel(ifName, 1); err != nil {
			return fmt.Errorf("error while initializing %s to channel 1: %s", ifName, err)
		}

		mod.Info("started (min rssi: %d dBm)", mod.minRSSI)
	}

	return nil
}

func (mod *WiFiModule) updateInfo(dot11 *layers.Dot11, packet gopacket.Packet) {
	// avoid parsing info from frames we're sending
	staMac := ops.Ternary(dot11.Flags.FromDS(), dot11.Address1, dot11.Address2).(net.HardwareAddr)
	if !bytes.Equal(staMac, mod.iface.HW) {
		if ok, enc, cipher, auth := packets.Dot11ParseEncryption(packet, dot11); ok {
			// Sometimes we get incomplete info about encryption, which
			// makes stations with encryption enabled switch to OPEN.
			// Prevent this behaviour by not downgrading the encryption.
			bssid := dot11.Address3.String()
			if station, found := mod.Session.WiFi.Get(bssid); found && station.IsOpen() {
				station.Encryption = enc
				station.Cipher = cipher
				station.Authentication = auth
			}
		}

		if ok, bssid, info := packets.Dot11ParseWPS(packet, dot11); ok {
			if station, found := mod.Session.WiFi.Get(bssid.String()); found {
				for name, value := range info {
					station.WPS[name] = value
				}
			}
		}
	}
}

func (mod *WiFiModule) updateStats(dot11 *layers.Dot11, packet gopacket.Packet) {
	// collect stats from data frames
	if dot11.Type.MainType() == layers.Dot11TypeData {
		bytes := uint64(len(packet.Data()))

		dst := dot11.Address1.String()
		if ap, found := mod.Session.WiFi.Get(dst); found {
			ap.Received += bytes
		} else if sta, found := mod.Session.WiFi.GetClient(dst); found {
			sta.Received += bytes
		}

		src := dot11.Address2.String()
		if ap, found := mod.Session.WiFi.Get(src); found {
			ap.Sent += bytes
		} else if sta, found := mod.Session.WiFi.GetClient(src); found {
			sta.Sent += bytes
		}
	}
}

func (mod *WiFiModule) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	mod.SetRunning(true, func() {
		// start channel hopper if needed
		if mod.channel == 0 && mod.source == "" {
			go mod.channelHopper()
		}

		// start the pruner
		go mod.stationPruner()

		mod.reads.Add(1)
		defer mod.reads.Done()

		src := gopacket.NewPacketSource(mod.handle, mod.handle.LinkType())
		mod.pktSourceChan = src.Packets()
		for packet := range mod.pktSourceChan {
			if !mod.Running() {
				break
			} else if packet == nil {
				continue
			}

			if mod.iface == mod.Session.Interface {
				mod.Session.Queue.TrackPacket(uint64(len(packet.Data())))
			}

			// perform initial dot11 parsing and layers validation
			if ok, radiotap, dot11 := packets.Dot11Parse(packet); ok {
				// check FCS checksum
				if mod.skipBroken && !dot11.ChecksumValid() {
					mod.Debug("skipping dot11 packet with invalid checksum.")
					continue
				}

				mod.discoverProbes(radiotap, dot11, packet)
				mod.discoverAccessPoints(radiotap, dot11, packet)
				mod.discoverClients(radiotap, dot11, packet)
				mod.discoverHandshakes(radiotap, dot11, packet)
				mod.discoverDeauths(radiotap, dot11, packet)
				mod.updateInfo(dot11, packet)
				mod.updateStats(dot11, packet)
			}
		}

		mod.pktSourceChanClosed = true
	})

	return nil
}

func (mod *WiFiModule) forcedStop() error {
	return mod.SetRunning(false, func() {
		// signal the main for loop we want to exit
		if !mod.pktSourceChanClosed {
			mod.pktSourceChan <- nil
		}
		// close the pcap handle to make the main for exit
		mod.handle.Close()
	})
}

func (mod *WiFiModule) Stop() error {
	return mod.SetRunning(false, func() {
		// wait any pending write operation
		mod.writes.Wait()
		// signal the main for loop we want to exit
		if !mod.pktSourceChanClosed {
			mod.pktSourceChan <- nil
		}
		mod.reads.Wait()
		// close the pcap handle to make the main for exit
		mod.handle.Close()
	})
}
