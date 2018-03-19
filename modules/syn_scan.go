package modules

import (
	"fmt"
	"net"
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

	"github.com/malfunkt/iprange"
)

const synSourcePort = 666

type SynScanner struct {
	session.SessionModule
	addresses []net.IP
	startPort int
	endPort   int
	waitGroup *sync.WaitGroup
}

func NewSynScanner(s *session.Session) *SynScanner {
	ss := &SynScanner{
		SessionModule: session.NewSessionModule("syn.scan", s),
		addresses:     make([]net.IP, 0),
		startPort:     0,
		endPort:       0,
		waitGroup:     &sync.WaitGroup{},
	}

	ss.AddHandler(session.NewModuleHandler("syn.scan IP-RANGE [START-PORT] [END-PORT]", "syn.scan ([^\\s]+) ?(\\d+)?([\\s\\d]*)?",
		"Perform a syn port scanning against an IP address within the provided ports range.",
		func(args []string) error {
			if ss.Running() == true {
				return fmt.Errorf("A scan is already running, wait for it to end before starting a new one.")
			}

			list, err := iprange.Parse(args[0])
			if err != nil {
				return fmt.Errorf("Error while parsing IP range '%s': %s", args[0], err)
			}

			argc := len(args)
			ss.addresses = list.Expand()
			ss.startPort = 1
			ss.endPort = 65535

			if argc > 1 && core.Trim(args[1]) != "" {
				if ss.startPort, err = strconv.Atoi(core.Trim(args[1])); err != nil {
					return fmt.Errorf("Invalid START-PORT: %s", err)
				}

				if ss.startPort > 65535 {
					ss.startPort = 65535
				}
				ss.endPort = ss.startPort
			}

			if argc > 2 && core.Trim(args[2]) != "" {
				if ss.endPort, err = strconv.Atoi(core.Trim(args[2])); err != nil {
					return fmt.Errorf("Invalid END-PORT: %s", err)
				}
			}

			if ss.endPort < ss.startPort {
				return fmt.Errorf("END-PORT is greater than START-PORT")
			}

			return ss.synScan()
		}))

	return ss
}

func (s *SynScanner) Name() string {
	return "syn.scan"
}

func (s *SynScanner) Description() string {
	return "A module to perform SYN port scanning."
}

func (s *SynScanner) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (s *SynScanner) Configure() error {
	return nil
}

func (s *SynScanner) Start() error {
	return nil
}

func (s *SynScanner) inRange(ip net.IP) bool {
	for _, a := range s.addresses {
		if a.Equal(ip) {
			return true
		}
	}
	return false
}

func (s *SynScanner) onPacket(pkt gopacket.Packet) {
	var eth layers.Ethernet
	var ip layers.IPv4
	var tcp layers.TCP
	foundLayerTypes := []gopacket.LayerType{}

	parser := gopacket.NewDecodingLayerParser(
		layers.LayerTypeEthernet,
		&eth,
		&ip,
		&tcp,
	)

	err := parser.DecodeLayers(pkt.Data(), &foundLayerTypes)
	if err != nil {
		return
	}

	if s.inRange(ip.SrcIP) && tcp.DstPort == synSourcePort && tcp.SYN && tcp.ACK {
		from := ip.SrcIP.String()
		port := int(tcp.SrcPort)

		var host *network.Endpoint
		if ip.SrcIP.Equal(s.Session.Interface.IP) {
			host = s.Session.Interface
		} else if ip.SrcIP.Equal(s.Session.Gateway.IP) {
			host = s.Session.Gateway
		} else {
			host = s.Session.Lan.GetByIp(from)
		}

		if host != nil {
			ports := host.Meta.GetIntsWith("tcp-ports", port, true)
			host.Meta.SetInts("tcp-ports", ports)
		}

		NewSynScanEvent(from, host, port).Push()
	}
}

func (s *SynScanner) synScan() error {
	s.SetRunning(true, func() {
		defer s.SetRunning(false, nil)

		s.waitGroup.Add(1)
		defer s.waitGroup.Done()

		naddrs := len(s.addresses)
		plural := "es"
		if naddrs == 1 {
			plural = ""
		}

		if s.startPort != s.endPort {
			log.Info("SYN scanning %d address%s from port %d to port %d ...", naddrs, plural, s.startPort, s.endPort)
		} else {
			log.Info("SYN scanning %d address%s on port %d ...", naddrs, plural, s.startPort)
		}

		// set the collector
		s.Session.Queue.OnPacket(s.onPacket)
		defer s.Session.Queue.OnPacket(nil)

		// start sending SYN packets and wait
		for _, address := range s.addresses {
			if s.Running() == false {
				break
			}
			mac, err := findMAC(s.Session, address, true)
			if err != nil {
				log.Debug("Could not get MAC for %s: %s", address.String(), err)
				continue
			}

			for dstPort := s.startPort; dstPort < s.endPort+1; dstPort++ {
				if s.Running() == false {
					break
				}

				err, raw := packets.NewTCPSyn(s.Session.Interface.IP, s.Session.Interface.HW, address, mac, synSourcePort, dstPort)
				if err != nil {
					log.Error("Error creating SYN packet: %s", err)
					continue
				}

				if err := s.Session.Queue.Send(raw); err != nil {
					log.Error("Error sending SYN packet: %s", err)
				} else {
					log.Debug("Sent %d bytes of SYN packet to %s for port %d", len(raw), address.String(), dstPort)
				}
			}
		}

		nports := s.endPort - s.startPort + 1
		time.Sleep(time.Duration(nports*500) * time.Millisecond)
	})

	return nil
}

func (s *SynScanner) Stop() error {
	return s.SetRunning(false, func() {
		s.waitGroup.Wait()
	})
}
