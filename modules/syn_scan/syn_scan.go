package syn_scan

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/async"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

const synSourcePort = 666

type synScannerStats struct {
	numPorts     uint64
	numAddresses uint64
	totProbes    uint64
	doneProbes   uint64
	openPorts    uint64
	started      time.Time
}

type SynScanner struct {
	session.SessionModule
	addresses     []net.IP
	startPort     int
	endPort       int
	handle        *pcap.Handle
	packets       chan gopacket.Packet
	progressEvery time.Duration
	stats         synScannerStats
	waitGroup     *sync.WaitGroup
	scanQueue     *async.WorkQueue
	bannerQueue   *async.WorkQueue
}

func NewSynScanner(s *session.Session) *SynScanner {
	mod := &SynScanner{
		SessionModule: session.NewSessionModule("syn.scan", s),
		addresses:     make([]net.IP, 0),
		waitGroup:     &sync.WaitGroup{},
		progressEvery: time.Duration(1) * time.Second,
	}

	mod.scanQueue = async.NewQueue(0, mod.scanWorker)
	mod.bannerQueue = async.NewQueue(0, mod.bannerGrabber)

	mod.State.Store("scanning", &mod.addresses)
	mod.State.Store("progress", 0.0)

	mod.AddParam(session.NewIntParameter("syn.scan.show-progress-every",
		"1",
		"Period in seconds for the scanning progress reporting."))

	mod.AddHandler(session.NewModuleHandler("syn.scan stop", "syn\\.scan (stop|off)",
		"Stop the current syn scanning session.",
		func(args []string) error {
			if !mod.Running() {
				return fmt.Errorf("no syn.scan is running")
			}
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("syn.scan IP-RANGE START-PORT END-PORT", "syn.scan ([^\\s]+) ?(\\d+)?([\\s\\d]*)?",
		"Perform a syn port scanning against an IP address within the provided ports range.",
		func(args []string) error {
			period := 0
			if mod.Running() {
				return fmt.Errorf("A scan is already running, wait for it to end before starting a new one.")
			} else if err := mod.parseTargets(args[0]); err != nil {
				return err
			} else if err = mod.parsePorts(args); err != nil {
				return err
			} else if err, period = mod.IntParam("syn.scan.show-progress-every"); err != nil {
				return err
			} else {
				mod.progressEvery = time.Duration(period) * time.Second
			}
			return mod.synScan()
		}))

	mod.AddHandler(session.NewModuleHandler("syn.scan.progress", "syn\\.scan\\.progress",
		"Print progress of the current syn scanning session.",
		func(args []string) error {
			if !mod.Running() {
				return fmt.Errorf("no syn.scan is running")
			}
			return mod.showProgress()
		}))

	return mod
}

func (mod *SynScanner) Name() string {
	return "syn.scan"
}

func (mod *SynScanner) Description() string {
	return "A module to perform SYN port scanning."
}

func (mod *SynScanner) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *SynScanner) Configure() (err error) {
	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	}
	if mod.handle == nil {
		if mod.handle, err = network.Capture(mod.Session.Interface.Name()); err != nil {
			return err
		} else if err = mod.handle.SetBPFFilter(fmt.Sprintf("tcp dst port %d", synSourcePort)); err != nil {
			return err
		}
		mod.packets = gopacket.NewPacketSource(mod.handle, mod.handle.LinkType()).Packets()
	}
	return nil
}

func (mod *SynScanner) Start() error {
	return nil
}

func plural(n uint64) string {
	if n > 1 {
		return "s"
	}
	return ""
}

func (mod *SynScanner) showProgress() error {
	progress := 100.0 * (float64(mod.stats.doneProbes) / float64(mod.stats.totProbes))
	mod.State.Store("progress", progress)
	mod.Info("[%.2f%%] found %d open port%s for %d address%s, sent %d/%d packets in %s",
		progress,
		mod.stats.openPorts,
		plural(mod.stats.openPorts),
		mod.stats.numAddresses,
		plural(mod.stats.numAddresses),
		mod.stats.doneProbes,
		mod.stats.totProbes,
		time.Since(mod.stats.started))
	return nil
}

func (mod *SynScanner) Stop() error {
	mod.Info("stopping ...")
	return mod.SetRunning(false, func() {
		mod.packets <- nil
		mod.waitGroup.Wait()
	})
}

type scanJob struct {
	Address net.IP
	Mac     net.HardwareAddr
}

func (mod *SynScanner) scanWorker(job async.Job) {
	scan := job.(scanJob)

	fromHW := mod.Session.Interface.HW
	fromIP := mod.Session.Interface.IP
	if scan.Address.To4() == nil {
		fromIP = mod.Session.Interface.IPv6
	}

	for dstPort := mod.startPort; dstPort < mod.endPort+1; dstPort++ {
		if !mod.Running() {
			break
		}

		atomic.AddUint64(&mod.stats.doneProbes, 1)

		err, raw := packets.NewTCPSyn(fromIP, fromHW, scan.Address, scan.Mac, synSourcePort, dstPort)
		if err != nil {
			mod.Error("error creating SYN packet: %s", err)
			continue
		}

		if err := mod.Session.Queue.Send(raw); err != nil {
			mod.Error("error sending SYN packet: %s", err)
		} else {
			mod.Debug("sent %d bytes of SYN packet to %s for port %d", len(raw), scan.Address.String(), dstPort)
		}

		time.Sleep(time.Duration(15) * time.Millisecond)
	}
}

func (mod *SynScanner) synScan() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	mod.SetRunning(true, func() {
		mod.waitGroup.Add(1)
		defer mod.waitGroup.Done()

		defer mod.SetRunning(false, func() {
			mod.showProgress()
			mod.addresses = []net.IP{}
			mod.State.Store("progress", 0.0)
			mod.State.Store("scanning", &mod.addresses)
			mod.packets <- nil
		})

		mod.stats.openPorts = 0
		mod.stats.numPorts = uint64(mod.endPort - mod.startPort + 1)
		mod.stats.started = time.Now()
		mod.stats.numAddresses = uint64(len(mod.addresses))
		mod.stats.totProbes = mod.stats.numAddresses * mod.stats.numPorts
		mod.stats.doneProbes = 0
		plural := "es"
		if mod.stats.numAddresses == 1 {
			plural = ""
		}

		if mod.stats.numPorts > 1 {
			mod.Info("scanning %d address%s from port %d to port %d ...", mod.stats.numAddresses, plural, mod.startPort, mod.endPort)
		} else {
			mod.Info("scanning %d address%s on port %d ...", mod.stats.numAddresses, plural, mod.startPort)
		}

		mod.State.Store("progress", 0.0)

		// start the collector
		mod.waitGroup.Add(1)
		go func() {
			defer mod.waitGroup.Done()

			for packet := range mod.packets {
				if !mod.Running() {
					break
				}
				mod.onPacket(packet)
			}
		}()

		// start to show progress every second
		go func() {
			for {
				time.Sleep(mod.progressEvery)
				if mod.Running() {
					mod.showProgress()
				} else {
					break
				}
			}
		}()

		// start sending SYN packets and wait
		for _, address := range mod.addresses {
			if !mod.Running() {
				break
			}
			mac, err := mod.Session.FindMAC(address, true)
			if err != nil {
				atomic.AddUint64(&mod.stats.doneProbes, mod.stats.numPorts)
				mod.Debug("could not get MAC for %s: %s", address.String(), err)
				continue
			}

			mod.scanQueue.Add(async.Job(scanJob{
				Address: address,
				Mac:     mac,
			}))
		}

		mod.scanQueue.WaitDone()
	})

	return nil
}
