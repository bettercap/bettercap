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
	mod := &Sniffer{
		SessionModule: session.NewSessionModule("net.sniff", s),
		Stats:         nil,
	}

	mod.SessionModule.Requires("net.recon")

	mod.AddParam(session.NewBoolParameter("net.sniff.verbose",
		"false",
		"If true, every captured and parsed packet will be sent to the events.stream for displaying, otherwise only the ones parsed at the application layer (sni, http, etc)."))

	mod.AddParam(session.NewBoolParameter("net.sniff.local",
		"false",
		"If true it will consider packets from/to this computer, otherwise it will skip them."))

	mod.AddParam(session.NewStringParameter("net.sniff.filter",
		"not arp",
		"",
		"BPF filter for the sniffer."))

	mod.AddParam(session.NewStringParameter("net.sniff.regexp",
		"",
		"",
		"If set, only packets matching this regular expression will be considered."))

	mod.AddParam(session.NewStringParameter("net.sniff.output",
		"",
		"",
		"If set, the sniffer will write captured packets to this file."))

	mod.AddParam(session.NewStringParameter("net.sniff.source",
		"",
		"",
		"If set, the sniffer will read from this pcap file instead of the current interface."))

	mod.AddHandler(session.NewModuleHandler("net.sniff stats", "",
		"Print sniffer session configuration and statistics.",
		func(args []string) error {
			if mod.Stats == nil {
				return fmt.Errorf("No stats yet.")
			}

			mod.Ctx.Log(mod.Session)

			return mod.Stats.Print()
		}))

	mod.AddHandler(session.NewModuleHandler("net.sniff on", "",
		"Start network sniffer in background.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("net.sniff off", "",
		"Stop network sniffer in background.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("net.fuzz on", "",
		"Enable fuzzing for every sniffed packet containing the specified layers.",
		func(args []string) error {
			return mod.StartFuzzing()
		}))

	mod.AddHandler(session.NewModuleHandler("net.fuzz off", "",
		"Disable fuzzing",
		func(args []string) error {
			return mod.StopFuzzing()
		}))

	mod.AddParam(session.NewStringParameter("net.fuzz.layers",
		"Payload",
		"",
		"Types of layer to fuzz."))

	mod.AddParam(session.NewDecimalParameter("net.fuzz.rate",
		"1.0",
		"Rate in the [0.0,1.0] interval of packets to fuzz."))

	mod.AddParam(session.NewDecimalParameter("net.fuzz.ratio",
		"0.4",
		"Rate in the [0.0,1.0] interval of bytes to fuzz for each packet."))

	mod.AddParam(session.NewBoolParameter("net.fuzz.silent",
		"false",
		"If true it will not report fuzzed packets."))

	return mod
}

func (mod Sniffer) Name() string {
	return "net.sniff"
}

func (mod Sniffer) Description() string {
	return "Sniff packets from the network."
}

func (mod Sniffer) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod Sniffer) isLocalPacket(packet gopacket.Packet) bool {
	ip4l := packet.Layer(layers.LayerTypeIPv4)
	if ip4l != nil {
		ip4, _ := ip4l.(*layers.IPv4)
		if ip4.SrcIP.Equal(mod.Session.Interface.IP) || ip4.DstIP.Equal(mod.Session.Interface.IP) {
			return true
		}
	} else {
		ip6l := packet.Layer(layers.LayerTypeIPv6)
		if ip6l != nil {
			ip6, _ := ip6l.(*layers.IPv6)
			if ip6.SrcIP.Equal(mod.Session.Interface.IPv6) || ip6.DstIP.Equal(mod.Session.Interface.IPv6) {
				return true
			}
		}
	}
	return false
}

func (mod *Sniffer) onPacketMatched(pkt gopacket.Packet) {
	if mainParser(pkt, mod.Ctx.Verbose) {
		mod.Stats.NumDumped++
	}
}

func (mod *Sniffer) Configure() error {
	var err error

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, mod.Ctx = mod.GetContext(); err != nil {
		if mod.Ctx != nil {
			mod.Ctx.Close()
			mod.Ctx = nil
		}
		return err
	}

	return nil
}

func (mod *Sniffer) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Stats = NewSnifferStats()

		src := gopacket.NewPacketSource(mod.Ctx.Handle, mod.Ctx.Handle.LinkType())
		mod.pktSourceChan = src.Packets()
		for packet := range mod.pktSourceChan {
			if !mod.Running() {
				mod.Debug("end pkt loop (pkt=%v filter='%s')", packet, mod.Ctx.Filter)
				break
			}

			now := time.Now()
			if mod.Stats.FirstPacket.IsZero() {
				mod.Stats.FirstPacket = now
			}
			mod.Stats.LastPacket = now

			isLocal := mod.isLocalPacket(packet)
			if isLocal {
				mod.Stats.NumLocal++
			}

			if mod.fuzzActive {
				mod.doFuzzing(packet)
			}

			if mod.Ctx.DumpLocal || !isLocal {
				data := packet.Data()
				if mod.Ctx.Compiled == nil || mod.Ctx.Compiled.Match(data) {
					mod.Stats.NumMatched++

					mod.onPacketMatched(packet)

					if mod.Ctx.OutputWriter != nil {
						mod.Ctx.OutputWriter.WritePacket(packet.Metadata().CaptureInfo, data)
						mod.Stats.NumWrote++
					}
				}
			}
		}

		mod.pktSourceChan = nil
	})
}

func (mod *Sniffer) Stop() error {
	return mod.SetRunning(false, func() {
		mod.Debug("stopping sniffer")
		if mod.pktSourceChan != nil {
			mod.Debug("sending nil")
			mod.pktSourceChan <- nil
			mod.Debug("nil sent")
		}
		mod.Debug("closing ctx")
		mod.Ctx.Close()
		mod.Debug("ctx closed")
	})
}
