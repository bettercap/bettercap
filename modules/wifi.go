package modules

import (
	"fmt"
	"net"
	"strconv"
	"strings"
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
)

type WiFiModule struct {
	session.SessionModule

	handle              *pcap.Handle
	source              string
	channel             int
	hopPeriod           time.Duration
	frequencies         []int
	ap                  *network.AccessPoint
	stickChan           int
	skipBroken          bool
	pktSourceChan       chan gopacket.Packet
	pktSourceChanClosed bool
	apRunning           bool
	apConfig            packets.Dot11ApConfig
	writes              *sync.WaitGroup
	reads               *sync.WaitGroup
}

func NewWiFiModule(s *session.Session) *WiFiModule {
	w := &WiFiModule{
		SessionModule: session.NewSessionModule("wifi", s),
		channel:       0,
		stickChan:     0,
		hopPeriod:     250 * time.Millisecond,
		ap:            nil,
		skipBroken:    true,
		apRunning:     false,
		writes:        &sync.WaitGroup{},
		reads:         &sync.WaitGroup{},
	}

	w.AddHandler(session.NewModuleHandler("wifi.recon on", "",
		"Start 802.11 wireless base stations discovery and channel hopping.",
		func(args []string) error {
			return w.Start()
		}))

	w.AddHandler(session.NewModuleHandler("wifi.recon off", "",
		"Stop 802.11 wireless base stations discovery and channel hopping.",
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
				w.stickChan = network.Dot11Freq2Chan(ap.Frequency)
				return nil
			}
			return fmt.Errorf("Could not find station with BSSID %s", args[0])
		}))

	w.AddHandler(session.NewModuleHandler("wifi.recon clear", "",
		"Remove the 802.11 base station filter.",
		func(args []string) error {
			w.ap = nil
			w.stickChan = 0
			var err error
			if w.frequencies, err = network.GetSupportedFrequencies(w.Session.Interface.Name()); err != nil {
				return err
			}
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

	w.AddHandler(session.NewModuleHandler("wifi.ap", "",
		"Inject fake management beacons in order to create a rogue access point.",
		func(args []string) error {
			if err := w.parseApConfig(); err != nil {
				return err
			} else {
				return w.startAp()
			}
		}))

	w.AddParam(session.NewStringParameter("wifi.ap.ssid",
		"FreeWiFi",
		"",
		"SSID of the fake access point."))

	w.AddParam(session.NewStringParameter("wifi.ap.bssid",
		session.ParamRandomMAC,
		"[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}",
		"BSSID of the fake access point."))

	w.AddParam(session.NewIntParameter("wifi.ap.channel",
		"1",
		"Channel of the fake access point."))

	w.AddParam(session.NewBoolParameter("wifi.ap.encryption",
		"true",
		"If true, the fake access point will use WPA2, otherwise it'll result as an open AP."))

	w.AddHandler(session.NewModuleHandler("wifi.show", "",
		"Show current wireless stations list (default sorting by essid).",
		func(args []string) error {
			return w.Show("rssi")
		}))

	w.AddHandler(session.NewModuleHandler("wifi.recon.channel", `wifi\.recon\.channel[\s]+([0-9]+(?:[, ]+[0-9]+)*|clear)`,
		"WiFi channels (comma separated) or 'clear' for channel hopping.",
		func(args []string) error {
			newfrequencies := w.frequencies[:0]

			if len(args) > 0 && args[0] != "clear" {
				channels := strings.Split(args[0], ",")
				for _, c := range channels {
					trimmed := strings.Trim(c, " ")
					channel, err := strconv.Atoi(trimmed)
					if err != nil {
						return err
					}
					newfrequencies = append(newfrequencies, network.Dot11Chan2Freq(channel))
				}
			} else {
				// No channels setted, retrieve frequencies supported by the card
				if frequencies, err := network.GetSupportedFrequencies(w.Session.Interface.Name()); err != nil {
					return err
				} else {
					newfrequencies = frequencies
				}
			}

			w.frequencies = newfrequencies

			return nil
		}))

	w.AddParam(session.NewStringParameter("wifi.source.file",
		"",
		"",
		"If set, the wifi module will read from this pcap file instead of the hardware interface."))

	w.AddParam(session.NewIntParameter("wifi.hop.period",
		"250",
		"If channel hopping is enabled (empty wifi.recon.channel), this is the time in milliseconds the algorithm will hop on every channel (it'll be doubled if both 2.4 and 5.0 bands are available)."))

	w.AddParam(session.NewBoolParameter("wifi.skip-broken",
		"true",
		"If true, dot11 packets with an invalid checksum will be skipped."))

	return w
}

func (w WiFiModule) Name() string {
	return "wifi"
}

func (w WiFiModule) Description() string {
	return "A module to monitor and perform wireless attacks on 802.11."
}

func (w WiFiModule) Author() string {
	return "Gianluca Braga <matrix86@protonmail.com> && Simone Margaritelli <evilsocket@protonmail.com>>"
}

func (w *WiFiModule) Configure() error {
	var hopPeriod int
	var err error

	if err, w.source = w.StringParam("wifi.source.file"); err != nil {
		return err
	}

	if w.source != "" {
		if w.handle, err = pcap.OpenOffline(w.source); err != nil {
			return err
		}
	} else {
		ihandle, err := pcap.NewInactiveHandle(w.Session.Interface.Name())
		if err != nil {
			return err
		}
		defer ihandle.CleanUp()

		if err = ihandle.SetRFMon(true); err != nil {
			return fmt.Errorf("Error while setting interface %s in monitor mode: %s", core.Bold(w.Session.Interface.Name()), err)
		} else if err = ihandle.SetSnapLen(65536); err != nil {
			return err
		} else if err = ihandle.SetTimeout(pcap.BlockForever); err != nil {
			return err
		} else if w.handle, err = ihandle.Activate(); err != nil {
			return err
		}
	}

	if err, w.skipBroken = w.BoolParam("wifi.skip-broken"); err != nil {
		return err
	} else if err, hopPeriod = w.IntParam("wifi.hop.period"); err != nil {
		return err
	}

	w.hopPeriod = time.Duration(hopPeriod) * time.Millisecond

	if w.source == "" {
		// No channels setted, retrieve frequencies supported by the card
		if len(w.frequencies) == 0 {
			if w.frequencies, err = network.GetSupportedFrequencies(w.Session.Interface.Name()); err != nil {
				return err
			}

			// we need to start somewhere, this is just to check if
			// this OS supports switching channel programmatically.
			if err = network.SetInterfaceChannel(w.Session.Interface.Name(), 1); err != nil {
				return err
			}
			log.Info("WiFi recon active with channel hopping.")
		}
	}

	return nil
}

func (w *WiFiModule) updateStats(dot11 *layers.Dot11, packet gopacket.Packet) {
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

func (w *WiFiModule) Start() error {
	if err := w.Configure(); err != nil {
		return err
	}

	w.SetRunning(true, func() {
		// start channel hopper if needed
		if w.channel == 0 && w.source == "" {
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
			} else if packet == nil {
				continue
			}

			w.Session.Queue.TrackPacket(uint64(len(packet.Data())))

			// perform initial dot11 parsing and layers validation
			if ok, radiotap, dot11 := packets.Dot11Parse(packet); ok == true {
				// check FCS checksum
				if w.skipBroken && dot11.ChecksumValid() == false {
					log.Debug("Skipping dot11 packet with invalid checksum.")
					continue
				}

				w.discoverProbes(radiotap, dot11, packet)
				w.discoverAccessPoints(radiotap, dot11, packet)
				w.discoverClients(radiotap, dot11, packet)
				w.updateStats(dot11, packet)
			}
		}
		w.pktSourceChanClosed = true
	})

	return nil
}

func (w *WiFiModule) Stop() error {
	return w.SetRunning(false, func() {
		// wait any pending write operation
		w.writes.Wait()
		// signal the main for loop we want to exit
		if w.pktSourceChanClosed == false {
			w.pktSourceChan <- nil
		}
		// close the pcap handle to make the main for exit
		w.handle.Close()
		// close the pcap handle to make the main for exit
		// wait for the loop to exit.
		w.reads.Wait()
	})
}
