package net_sniff

import (
	"fmt"
	"time"

	"github.com/bettercap/bettercap/session"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Sniffer struct {
	session.SessionModule
	Stats         *SnifferStats
	Ctx           *SnifferContext
	pktSourceChan chan gopacket.Packet

	fuzzActive bool
	fuzzSilent bool
	fuzzLayers []string
	fuzzRate   float64
	fuzzRatio  float64
}

func NewSniffer(s *session.Session) *Sniffer {
	sniff := &Sniffer{
		SessionModule: session.NewSessionModule("net.sniff", s),
		Stats:         nil,
	}

	sniff.AddParam(session.NewBoolParameter("net.sniff.verbose",
		"false",
		"If true, every captured and parsed packet will be sent to the events.stream for displaying, otherwise only the ones parsed at the application layer (sni, http, etc)."))

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
		"If set, only packets matching this regular expression will be considered."))

	sniff.AddParam(session.NewStringParameter("net.sniff.output",
		"",
		"",
		"If set, the sniffer will write captured packets to this file."))

	sniff.AddParam(session.NewStringParameter("net.sniff.source",
		"",
		"",
		"If set, the sniffer will read from this pcap file instead of the current interface."))

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

	sniff.AddHandler(session.NewModuleHandler("net.fuzz on", "",
		"Enable fuzzing for every sniffed packet containing the sapecified layers.",
		func(args []string) error {
			return sniff.StartFuzzing()
		}))

	sniff.AddHandler(session.NewModuleHandler("net.fuzz off", "",
		"Disable fuzzing",
		func(args []string) error {
			return sniff.StopFuzzing()
		}))

	sniff.AddParam(session.NewStringParameter("net.fuzz.layers",
		"Payload",
		"",
		"Types of layer to fuzz."))

	sniff.AddParam(session.NewDecimalParameter("net.fuzz.rate",
		"1.0",
		"Rate in the [0.0,1.0] interval of packets to fuzz."))

	sniff.AddParam(session.NewDecimalParameter("net.fuzz.ratio",
		"0.4",
		"Rate in the [0.0,1.0] interval of bytes to fuzz for each packet."))

	sniff.AddParam(session.NewBoolParameter("net.fuzz.silent",
		"false",
		"If true it will not report fuzzed packets."))

	return sniff
}

func (s Sniffer) Name() string {
	return "net.sniff"
}

func (s Sniffer) Description() string {
	return "Sniff packets from the network."
}

func (s Sniffer) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (s Sniffer) isLocalPacket(packet gopacket.Packet) bool {
	ipl := packet.Layer(layers.LayerTypeIPv4)
	if ipl != nil {
		ip, _ := ipl.(*layers.IPv4)
		if ip.SrcIP.Equal(s.Session.Interface.IP) || ip.DstIP.Equal(s.Session.Interface.IP) {
			return true
		}
	}
	return false
}

func (s *Sniffer) onPacketMatched(pkt gopacket.Packet) {
	if mainParser(pkt, s.Ctx.Verbose) {
		s.Stats.NumDumped++
	}
}

func (s *Sniffer) Configure() error {
	var err error

	if s.Running() {
		return session.ErrAlreadyStarted
	} else if err, s.Ctx = s.GetContext(); err != nil {
		if s.Ctx != nil {
			s.Ctx.Close()
			s.Ctx = nil
		}
		return err
	}

	return nil
}

func (s *Sniffer) Start() error {
	if err := s.Configure(); err != nil {
		return err
	}

	return s.SetRunning(true, func() {
		s.Stats = NewSnifferStats()

		src := gopacket.NewPacketSource(s.Ctx.Handle, s.Ctx.Handle.LinkType())
		s.pktSourceChan = src.Packets()
		for packet := range s.pktSourceChan {
			if !s.Running() {
				s.Debug("end pkt loop (pkt=%v filter='%s')", packet, s.Ctx.Filter)
				break
			}

			now := time.Now()
			if s.Stats.FirstPacket.IsZero() {
				s.Stats.FirstPacket = now
			}
			s.Stats.LastPacket = now

			isLocal := s.isLocalPacket(packet)
			if isLocal {
				s.Stats.NumLocal++
			}

			if s.fuzzActive {
				s.doFuzzing(packet)
			}

			if s.Ctx.DumpLocal || !isLocal {
				data := packet.Data()
				if s.Ctx.Compiled == nil || s.Ctx.Compiled.Match(data) {
					s.Stats.NumMatched++

					s.onPacketMatched(packet)

					if s.Ctx.OutputWriter != nil {
						s.Ctx.OutputWriter.WritePacket(packet.Metadata().CaptureInfo, data)
						s.Stats.NumWrote++
					}
				}
			}
		}

		s.pktSourceChan = nil
	})
}

func (s *Sniffer) Stop() error {
	return s.SetRunning(false, func() {
		s.Debug("stopping sniffer")
		if s.pktSourceChan != nil {
			s.Debug("sending nil")
			s.pktSourceChan <- nil
			s.Debug("nil sent")
		}
		s.Debug("closing ctx")
		s.Ctx.Close()
		s.Debug("ctx closed")
	})
}
