package session_modules

import (
	"fmt"
	"github.com/evilsocket/bettercap/core"
	"github.com/evilsocket/bettercap/session"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"net"
	"os"
	"regexp"
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

func (c *SnifferContext) Log() {
	log.Info("\n")

	if c.DumpLocal {
		log.Info("  Skip local packets : " + no)
	} else {
		log.Info("  Skip local packets : " + yes)
	}

	if c.Verbose {
		log.Info("  Verbose            : " + yes)
	} else {
		log.Info("  Verbose            : " + no)
	}

	if c.Filter != "" {
		log.Info("  BPF Filter         : '" + core.Yellow(c.Filter) + "'")
	}

	if c.Expression != "" {
		log.Info("  Regular expression : '" + core.Yellow(c.Expression) + "'")
	}

	if c.Output != "" {
		log.Info("  File output        : '" + core.Yellow(c.Output) + "'")
	}

	log.Info("\n")
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

type Sniffer struct {
	session.SessionModule
}

func NewSniffer(s *session.Session) *Sniffer {
	sniff := &Sniffer{
		SessionModule: session.NewSessionModule(s),
	}

	sniff.AddParam(session.NewBoolParameter("net.sniffer.verbose", "true", "", "Print captured packets to screen."))
	sniff.AddParam(session.NewBoolParameter("net.sniffer.local", "false", "", "If true it will consider packets from/to this computer, otherwise it will skip them."))
	sniff.AddParam(session.NewStringParameter("net.sniffer.filter", "not arp", "", "BPF filter for the sniffer."))
	sniff.AddParam(session.NewStringParameter("net.sniffer.regexp", "", "", "If filled, only packets matching this regular expression will be considered."))
	sniff.AddParam(session.NewStringParameter("net.sniffer.output", "", "", "If set, the sniffer will write captured packets to this file."))

	sniff.AddHandler(session.NewModuleHandler("net.sniffer (on|off)", "^net\\.sniffer\\s+(on|off)$",
		"Start/stop network sniffer in background.",
		func(args []string) error {
			if args[0] == "on" {
				return sniff.Start()
			} else {
				return sniff.Stop()
			}
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

func (s *Sniffer) Start() error {
	if s.Running() == false {
		var err error
		var ctx *SnifferContext

		if err, ctx = s.GetContext(); err != nil {
			if ctx != nil {
				ctx.Close()
			}
			return err
		}

		s.SetRunning(true)

		go func() {
			defer ctx.Close()

			log.Info("Network sniffer started.\n")
			ctx.Log()

			src := gopacket.NewPacketSource(ctx.Handle, ctx.Handle.LinkType())
			for packet := range src.Packets() {
				if s.Running() == false {
					break
				}

				if ctx.DumpLocal == true || s.isLocalPacket(packet) == false {
					data := packet.Data()
					if ctx.Compiled == nil || ctx.Compiled.Match(data) == true {
						if ctx.Verbose {
							fmt.Println(packet.Dump())
						}

						if ctx.OutputWriter != nil {
							ctx.OutputWriter.WritePacket(packet.Metadata().CaptureInfo, data)
						}
					}
				}
			}

			log.Info("Network sniffer stopped.\n")
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
