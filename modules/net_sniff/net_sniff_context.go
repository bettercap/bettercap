package net_sniff

import (
	"os"
	"regexp"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"

	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"

	"github.com/evilsocket/islazy/tui"
)

type SnifferContext struct {
	Handle       *pcap.Handle
	Source       string
	DumpLocal    bool
	Verbose      bool
	Filter       string
	Expression   string
	Compiled     *regexp.Regexp
	Output       string
	OutputFile   *os.File
	OutputWriter *pcapgo.Writer
}

func (mod *Sniffer) GetContext() (error, *SnifferContext) {
	var err error

	ctx := NewSnifferContext()

	if err, ctx.Source = mod.StringParam("net.sniff.source"); err != nil {
		return err, ctx
	}

	if ctx.Source == "" {
		/*
		 * We don't want to pcap.BlockForever otherwise pcap_close(handle)
		 * could hang waiting for a timeout to expire ...
		 */
		readTimeout := 500 * time.Millisecond
		if ctx.Handle, err = network.CaptureWithTimeout(mod.Session.Interface.Name(), readTimeout); err != nil {
			return err, ctx
		}
	} else {
		if ctx.Handle, err = pcap.OpenOffline(ctx.Source); err != nil {
			return err, ctx
		}
	}

	if err, ctx.Verbose = mod.BoolParam("net.sniff.verbose"); err != nil {
		return err, ctx
	}

	if err, ctx.DumpLocal = mod.BoolParam("net.sniff.local"); err != nil {
		return err, ctx
	}

	if err, ctx.Filter = mod.StringParam("net.sniff.filter"); err != nil {
		return err, ctx
	} else if ctx.Filter != "" {
		err = ctx.Handle.SetBPFFilter(ctx.Filter)
		if err != nil {
			return err, ctx
		}
	}

	if err, ctx.Expression = mod.StringParam("net.sniff.regexp"); err != nil {
		return err, ctx
	} else if ctx.Expression != "" {
		if ctx.Compiled, err = regexp.Compile(ctx.Expression); err != nil {
			return err, ctx
		}
	}

	if err, ctx.Output = mod.StringParam("net.sniff.output"); err != nil {
		return err, ctx
	} else if ctx.Output != "" {
		if ctx.OutputFile, err = os.Create(ctx.Output); err != nil {
			return err, ctx
		}

		ctx.OutputWriter = pcapgo.NewWriter(ctx.OutputFile)
		ctx.OutputWriter.WriteFileHeader(65536, ctx.Handle.LinkType())
	}

	return nil, ctx
}

func NewSnifferContext() *SnifferContext {
	return &SnifferContext{
		Handle:       nil,
		DumpLocal:    false,
		Verbose:      false,
		Filter:       "",
		Expression:   "",
		Compiled:     nil,
		Output:       "",
		OutputFile:   nil,
		OutputWriter: nil,
	}
}

var (
	no  = tui.Red("no")
	yes = tui.Green("yes")
	yn  = map[bool]string{
		true:  yes,
		false: no,
	}
)

func (c *SnifferContext) Log(sess *session.Session) {
	log.Info("Skip local packets : %s", yn[c.DumpLocal])
	log.Info("Verbose            : %s", yn[c.Verbose])
	log.Info("BPF Filter         : '%s'", tui.Yellow(c.Filter))
	log.Info("Regular expression : '%s'", tui.Yellow(c.Expression))
	log.Info("File output        : '%s'", tui.Yellow(c.Output))
}

func (c *SnifferContext) Close() {
	if c.Handle != nil {
		log.Debug("closing handle")
		c.Handle.Close()
		log.Debug("handle closed")
		c.Handle = nil
	}

	if c.OutputFile != nil {
		log.Debug("closing output")
		c.OutputFile.Close()
		log.Debug("output closed")
		c.OutputFile = nil
	}
}
