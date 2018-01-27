package modules

import (
	"fmt"
	"net"
	"time"

	"github.com/evilsocket/bettercap-ng/session"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Sniffer struct {
	session.SessionModule
	Stats *SnifferStats
	Ctx   *SnifferContext
}

func NewSniffer(s *session.Session) *Sniffer {
	sniff := &Sniffer{
		SessionModule: session.NewSessionModule("net.sniff", s),
		Stats:         nil,
	}

	sniff.AddParam(session.NewBoolParameter("net.sniff.verbose",
		"true",
		"If true, will print every captured packet, otherwise only selected ones."))

	sniff.AddParam(session.NewBoolParameter("net.sniff.local",
		"false",
		"If true it will consider packets from/to this computer, otherwise it will skip them."))

	sniff.AddParam(session.NewStringParameter("net.sniff.filter",
		"not arp",
		"",
		"BPF filter for the sniffer."))

	sniff.AddParam(session.NewStringParameter("net.sniff.regexp",
		"",
		"",
		"If filled, only packets matching this regular expression will be considered."))

	sniff.AddParam(session.NewStringParameter("net.sniff.output",
		"",
		"",
		"If set, the sniffer will write captured packets to this file."))

	sniff.AddHandler(session.NewModuleHandler("net.sniff stats", "",
		"Print sniffer session configuration and statistics.",
		func(args []string) error {
			if sniff.Stats == nil {
				return fmt.Errorf("No stats yet.")
			}

			sniff.Ctx.Log(sniff.Session)

			return sniff.Stats.Print()
		}))

	sniff.AddHandler(session.NewModuleHandler("net.sniff on", "",
		"Start network sniffer in background.",
		func(args []string) error {
			return sniff.Start()
		}))

	sniff.AddHandler(session.NewModuleHandler("net.sniff off", "",
		"Stop network sniffer in background.",
		func(args []string) error {
			return sniff.Stop()
		}))

	return sniff
}

func (s Sniffer) Name() string {
	return "net.sniff"
}

func (s Sniffer) Description() string {
	return "Sniff packets from the network."
}

func (s Sniffer) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func same(a, b net.HardwareAddr) bool {
	if len(a) != len(b) {
		return false
	}

	for idx, v := range a {
		if b[idx] != v {
			return false
		}
	}

	return true
}

func (s Sniffer) isLocalPacket(packet gopacket.Packet) bool {
	local_hw := s.Session.Interface.HW
	eth := packet.Layer(layers.LayerTypeEthernet)
	if eth != nil {
		eth_packet, _ := eth.(*layers.Ethernet)
		if same(eth_packet.SrcMAC, local_hw) || same(eth_packet.DstMAC, local_hw) {
			return true
		}
	}
	return false
}

func (s *Sniffer) onPacketMatched(pkt gopacket.Packet) {
	if mainParser(pkt, s.Ctx.Verbose) == true {
		s.Stats.NumDumped++
	}
}

func (s *Sniffer) Configure() error {
	var err error

	if err, s.Ctx = s.GetContext(); err != nil {
		if s.Ctx != nil {
			s.Ctx.Close()
			s.Ctx = nil
		}
		return err
	}

	return nil
}

func (s *Sniffer) Start() error {
	if s.Running() == true {
		return session.ErrAlreadyStarted
	} else if err := s.Configure(); err != nil {
		return err
	}

	s.SetRunning(true)

	go func() {
		s.Stats = NewSnifferStats()
		defer s.Ctx.Close()

		src := gopacket.NewPacketSource(s.Ctx.Handle, s.Ctx.Handle.LinkType())
		for packet := range src.Packets() {
			if s.Running() == false {
				break
			}

			now := time.Now()
			if s.Stats.FirstPacket.IsZero() {
				s.Stats.FirstPacket = now
			}
			s.Stats.LastPacket = now

			is_local := false
			if s.isLocalPacket(packet) {
				is_local = true
				s.Stats.NumLocal++
			}

			if s.Ctx.DumpLocal == true || is_local == false {
				data := packet.Data()
				if s.Ctx.Compiled == nil || s.Ctx.Compiled.Match(data) == true {
					s.Stats.NumMatched++

					s.onPacketMatched(packet)

					if s.Ctx.OutputWriter != nil {
						s.Ctx.OutputWriter.WritePacket(packet.Metadata().CaptureInfo, data)
						s.Stats.NumWrote++
					}
				}
			}
		}
	}()

	return nil
}

func (s *Sniffer) Stop() error {
	if s.Running() == false {
		return session.ErrAlreadyStopped
	}
	s.SetRunning(false)

	return nil
}
