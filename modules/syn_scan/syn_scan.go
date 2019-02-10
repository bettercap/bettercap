package syn_scan

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/malfunkt/iprange"

	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"
)

const synSourcePort = 666

type synScannerStats struct {
	started      time.Time
	numPorts     uint64
	numAddresses uint64
	totProbes    uint64
	doneProbes   uint64
	openPorts    uint64
}

type SynScanner struct {
	session.SessionModule
	addresses     []net.IP
	startPort     int
	endPort       int
	progressEvery time.Duration
	stats         synScannerStats
	waitGroup     *sync.WaitGroup
}

func NewSynScanner(s *session.Session) *SynScanner {
	ss := &SynScanner{
		SessionModule: session.NewSessionModule("syn.scan", s),
		addresses:     make([]net.IP, 0),
		waitGroup:     &sync.WaitGroup{},
		progressEvery: time.Duration(1) * time.Second,
	}

	ss.AddParam(session.NewIntParameter("syn.scan.show-progress-every",
		"1",
		"Period in seconds for the scanning progress reporting."))

	ss.AddHandler(session.NewModuleHandler("syn.scan stop", "syn\\.scan (stop|off)",
		"Stop the current syn scanning session.",
		func(args []string) error {
			if !ss.Running() {
				return fmt.Errorf("no syn.scan is running")
			}
			return ss.Stop()
		}))

	ss.AddHandler(session.NewModuleHandler("syn.scan IP-RANGE [START-PORT] [END-PORT]", "syn.scan ([^\\s]+) ?(\\d+)?([\\s\\d]*)?",
		"Perform a syn port scanning against an IP address within the provided ports range.",
		func(args []string) error {
			period := 0
			if ss.Running() {
				return fmt.Errorf("A scan is already running, wait for it to end before starting a new one.")
			} else if err := ss.parseTargets(args[0]); err != nil {
				return err
			} else if err = ss.parsePorts(args); err != nil {
				return err
			} else if err, period = ss.IntParam("syn.scan.show-progress-every"); err != nil {
				return err
			} else {
				ss.progressEvery = time.Duration(period) * time.Second
			}
			return ss.synScan()
		}))

	ss.AddHandler(session.NewModuleHandler("syn.scan.progress", "syn\\.scan\\.progress",
		"Print progress of the current syn scanning session.",
		func(args []string) error {
			if !ss.Running() {
				return fmt.Errorf("no syn.scan is running")
			}
			return ss.showProgress()
		}))

	return ss
}

func (s *SynScanner) parseTargets(arg string) error {
	if list, err := iprange.Parse(arg); err != nil {
		return fmt.Errorf("error while parsing IP range '%s': %s", arg, err)
	} else {
		s.addresses = list.Expand()
	}
	return nil
}

func (s *SynScanner) parsePorts(args []string) (err error) {
	argc := len(args)
	s.stats.totProbes = 0
	s.stats.doneProbes = 0
	s.startPort = 1
	s.endPort = 65535

	if argc > 1 && str.Trim(args[1]) != "" {
		if s.startPort, err = strconv.Atoi(str.Trim(args[1])); err != nil {
			return fmt.Errorf("invalid start port %s: %s", args[1], err)
		} else if s.startPort > 65535 {
			s.startPort = 65535
		}
		s.endPort = s.startPort
	}

	if argc > 2 && str.Trim(args[2]) != "" {
		if s.endPort, err = strconv.Atoi(str.Trim(args[2])); err != nil {
			return fmt.Errorf("invalid end port %s: %s", args[2], err)
		}
	}

	if s.endPort < s.startPort {
		return fmt.Errorf("end port %d is greater than start port %d", s.endPort, s.startPort)
	}

	return
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

func plural(n uint64) string {
	if n > 1 {
		return "s"
	}
	return ""
}

func (s *SynScanner) showProgress() error {
	progress := 100.0 * (float64(s.stats.doneProbes) / float64(s.stats.totProbes))
	log.Info("[%s] [%.2f%%] found %d open port%s for %d address%s, sent %d/%d packets in %s",
		tui.Green("syn.scan"),
		progress,
		s.stats.openPorts,
		plural(s.stats.openPorts),
		s.stats.numAddresses,
		plural(s.stats.numAddresses),
		s.stats.doneProbes,
		s.stats.totProbes,
		time.Since(s.stats.started))
	return nil
}

func (s *SynScanner) Stop() error {
	log.Info("[%s] stopping ...", tui.Green("syn.scan"))
	return s.SetRunning(false, func() {
		s.waitGroup.Wait()
		s.showProgress()
	})
}

func (s *SynScanner) synScan() error {
	s.SetRunning(true, func() {
		defer s.SetRunning(false, nil)

		s.waitGroup.Add(1)
		defer s.waitGroup.Done()

		s.stats.openPorts = 0
		s.stats.numPorts = uint64(s.endPort - s.startPort + 1)
		s.stats.started = time.Now()
		s.stats.numAddresses = uint64(len(s.addresses))
		s.stats.totProbes = s.stats.numAddresses * s.stats.numPorts
		s.stats.doneProbes = 0
		plural := "es"
		if s.stats.numAddresses == 1 {
			plural = ""
		}

		if s.stats.numPorts > 1 {
			log.Info("scanning %d address%s from port %d to port %d ...", s.stats.numAddresses, plural, s.startPort, s.endPort)
		} else {
			log.Info("scanning %d address%s on port %d ...", s.stats.numAddresses, plural, s.startPort)
		}

		// set the collector
		s.Session.Queue.OnPacket(s.onPacket)
		defer s.Session.Queue.OnPacket(nil)

		// start to show progress every second
		go func() {
			for {
				time.Sleep(s.progressEvery)
				if s.Running() {
					s.showProgress()
				} else {
					break
				}
			}
		}()

		// start sending SYN packets and wait
		for _, address := range s.addresses {
			if !s.Running() {
				break
			}
			mac, err := s.Session.FindMAC(address, true)
			if err != nil {
				atomic.AddUint64(&s.stats.doneProbes, s.stats.numPorts)
				log.Debug("Could not get MAC for %s: %s", address.String(), err)
				continue
			}

			for dstPort := s.startPort; dstPort < s.endPort+1; dstPort++ {
				if !s.Running() {
					break
				}

				atomic.AddUint64(&s.stats.doneProbes, 1)

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

				time.Sleep(time.Duration(10) * time.Millisecond)
			}
		}
	})

	return nil
}
