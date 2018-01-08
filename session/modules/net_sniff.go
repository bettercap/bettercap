package session_modules

import (
	"fmt"
	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/session"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"net"
	"os"
	"regexp"
	"time"
)

type SnifferContext struct {
	Handle       *pcap.Handle
	DumpLocal    bool
	Verbose      bool
	Filter       string
	Expression   string
	Compiled     *regexp.Regexp
	Output       string
	OutputFile   *os.File
	OutputWriter *pcapgo.Writer
}

func NewSnifferContext() *SnifferContext {
	return &SnifferContext{
		Handle:       nil,
		DumpLocal:    false,
		Verbose:      true,
		Filter:       "",
		Expression:   "",
		Compiled:     nil,
		Output:       "",
		OutputFile:   nil,
		OutputWriter: nil,
	}
}

var (
	no  = core.Red("no")
	yes = core.Green("yes")
)

func (c *SnifferContext) Log(sess *session.Session) {
	if c.DumpLocal {
		sess.Events.Log(session.INFO, "Skip local packets : "+no)
	} else {
		sess.Events.Log(session.INFO, "Skip local packets : "+yes)
	}

	if c.Verbose {
		sess.Events.Log(session.INFO, "Verbose            : "+yes)
	} else {
		sess.Events.Log(session.INFO, "Verbose            : "+no)
	}

	if c.Filter != "" {
		sess.Events.Log(session.INFO, "BPF Filter         : '"+core.Yellow(c.Filter)+"'")
	}

	if c.Expression != "" {
		sess.Events.Log(session.INFO, "Regular expression : '"+core.Yellow(c.Expression)+"'")
	}

	if c.Output != "" {
		sess.Events.Log(session.INFO, "File output        : '"+core.Yellow(c.Output)+"'")
	}
}

func (c *SnifferContext) Close() {
	if c.Handle != nil {
		c.Handle.Close()
		c.Handle = nil
	}

	if c.OutputFile != nil {
		c.OutputFile.Close()
		c.OutputFile = nil
	}
}

type SnifferStats struct {
	NumLocal    uint64
	NumMatched  uint64
	NumDumped   uint64
	NumWrote    uint64
	Started     time.Time
	FirstPacket time.Time
	LastPacket  time.Time
}

func NewSnifferStats() *SnifferStats {
	return &SnifferStats{
		NumLocal:    0,
		NumMatched:  0,
		NumDumped:   0,
		NumWrote:    0,
		Started:     time.Now(),
		FirstPacket: time.Time{},
		LastPacket:  time.Time{},
	}
}

type Sniffer struct {
	session.SessionModule
	Stats *SnifferStats
	Ctx   *SnifferContext
}

func NewSniffer(s *session.Session) *Sniffer {
	sniff := &Sniffer{
		SessionModule: session.NewSessionModule("net.sniffer", s),
		Stats:         nil,
	}

	sniff.AddParam(session.NewBoolParameter("net.sniffer.verbose",
		"true",
		"Print captured packets to screen."))

	sniff.AddParam(session.NewBoolParameter("net.sniffer.local",
		"false",
		"If true it will consider packets from/to this computer, otherwise it will skip them."))

	sniff.AddParam(session.NewStringParameter("net.sniffer.filter",
		"not arp",
		"",
		"BPF filter for the sniffer."))

	sniff.AddParam(session.NewStringParameter("net.sniffer.regexp",
		"",
		"",
		"If filled, only packets matching this regular expression will be considered."))

	sniff.AddParam(session.NewStringParameter("net.sniffer.output",
		"",
		"",
		"If set, the sniffer will write captured packets to this file."))

	sniff.AddHandler(session.NewModuleHandler("net.sniffer stats", "",
		"Print sniffer session configuration and statistics.",
		func(args []string) error {
			return sniff.PrintStats()
		}))

	sniff.AddHandler(session.NewModuleHandler("net.sniffer on", "",
		"Start network sniffer in background.",
		func(args []string) error {
			return sniff.Start()
		}))

	sniff.AddHandler(session.NewModuleHandler("net.sniffer off", "",
		"Stop network sniffer in background.",
		func(args []string) error {
			return sniff.Stop()
		}))

	return sniff
}

func (s Sniffer) Name() string {
	return "Network Sniffer"
}

func (s Sniffer) Description() string {
	return "Sniff packets from the network."
}

func (s Sniffer) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (sn Sniffer) OnSessionEnded(s *session.Session) {
	if sn.Running() {
		sn.Stop()
	}
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

func (s *Sniffer) GetContext() (error, *SnifferContext) {
	var err error

	ctx := NewSnifferContext()

	if ctx.Handle, err = pcap.OpenLive(s.Session.Interface.Name(), 65536, true, pcap.BlockForever); err != nil {
		return err, ctx
	}

	if err, v := s.Param("net.sniffer.verbose").Get(s.Session); err != nil {
		return err, ctx
	} else {
		ctx.Verbose = v.(bool)
	}

	if err, v := s.Param("net.sniffer.local").Get(s.Session); err != nil {
		return err, ctx
	} else {
		ctx.DumpLocal = v.(bool)
	}

	if err, v := s.Param("net.sniffer.filter").Get(s.Session); err != nil {
		return err, ctx
	} else {
		if ctx.Filter = v.(string); ctx.Filter != "" {
			err = ctx.Handle.SetBPFFilter(ctx.Filter)
			if err != nil {
				return err, ctx
			}
		}
	}

	if err, v := s.Param("net.sniffer.regexp").Get(s.Session); err != nil {
		return err, ctx
	} else {
		if ctx.Expression = v.(string); ctx.Expression != "" {
			if ctx.Compiled, err = regexp.Compile(ctx.Expression); err != nil {
				return err, ctx
			}
		}
	}

	if err, v := s.Param("net.sniffer.output").Get(s.Session); err != nil {
		return err, ctx
	} else {
		if ctx.Output = v.(string); ctx.Output != "" {
			if ctx.OutputFile, err = os.Create(ctx.Output); err != nil {
				return err, ctx
			}

			ctx.OutputWriter = pcapgo.NewWriter(ctx.OutputFile)
			ctx.OutputWriter.WriteFileHeader(65536, layers.LinkTypeEthernet)
		}
	}

	return nil, ctx
}

func (s *Sniffer) PrintStats() error {
	if s.Stats == nil {
		return fmt.Errorf("No stats yet.")
	}

	s.Ctx.Log(s.Session)

	first := "never"
	last := "never"

	if s.Stats.FirstPacket.IsZero() == false {
		first = s.Stats.FirstPacket.String()
	}

	if s.Stats.LastPacket.IsZero() == false {
		last = s.Stats.LastPacket.String()
	}

	s.Session.Events.Log(session.INFO, "Sniffer Started    : %s", s.Stats.Started)
	s.Session.Events.Log(session.INFO, "First Packet Seen  : %s", first)
	s.Session.Events.Log(session.INFO, "Last Packet Seen   : %s", last)
	s.Session.Events.Log(session.INFO, "Local Packets      : %d", s.Stats.NumLocal)
	s.Session.Events.Log(session.INFO, "Matched Packets    : %d", s.Stats.NumMatched)
	s.Session.Events.Log(session.INFO, "Dumped Packets     : %d", s.Stats.NumDumped)
	s.Session.Events.Log(session.INFO, "Wrote Packets      : %d", s.Stats.NumWrote)

	return nil
}

func (s *Sniffer) Start() error {
	if s.Running() == false {
		var err error

		if err, s.Ctx = s.GetContext(); err != nil {
			if s.Ctx != nil {
				s.Ctx.Close()
				s.Ctx = nil
			}
			return err
		}

		s.Stats = NewSnifferStats()
		s.SetRunning(true)

		go func() {
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

						if s.Ctx.Verbose {
							fmt.Println(packet.Dump())
							s.Stats.NumDumped++
						}

						if s.Ctx.OutputWriter != nil {
							s.Ctx.OutputWriter.WritePacket(packet.Metadata().CaptureInfo, data)
							s.Stats.NumWrote++
						}
					}
				}
			}
		}()

		return nil
	} else {
		return fmt.Errorf("Network sniffer already started.")
	}
}

func (s *Sniffer) Stop() error {
	if s.Running() == true {
		s.SetRunning(false)
		return nil
	} else {
		return fmt.Errorf("Network sniffer already stopped.")
	}
}
