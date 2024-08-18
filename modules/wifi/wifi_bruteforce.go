package wifi

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/bettercap/bettercap/v2/network"
	"github.com/evilsocket/islazy/async"
	"github.com/evilsocket/islazy/ops"
	"github.com/evilsocket/islazy/str"
)

var (
	errRecon          = errors.New("turn off wifi.recon first")
	errAlreadyRunning = errors.New("bruteforce already running")
	errNotRunning     = errors.New("bruteforce not running")
)

type bruteforceJob struct {
	running  *atomic.Bool
	done     *atomic.Uint64
	iface    string
	essid    string
	password string
	timeout  time.Duration
}

type BruteforceSuccess struct {
	Iface    string
	Target   string
	Password string
	Elapsed  time.Duration
}

type bruteforceConfig struct {
	running       atomic.Bool
	queue         *async.WorkQueue
	done          atomic.Uint64
	todo          uint64
	target        string
	wordlist      string
	workers       int
	timeout       int
	wide          bool
	stop_at_first bool

	passwords []string
	targets   []string
}

func NewBruteForceConfig() *bruteforceConfig {
	return &bruteforceConfig{
		wordlist:      "/usr/share/dict/words",
		passwords:     make([]string, 0),
		targets:       make([]string, 0),
		workers:       1,
		wide:          false,
		stop_at_first: true,
		timeout:       15,
		queue:         nil,
		done:          atomic.Uint64{},
		todo:          0,
	}
}

func (bruteforce *bruteforceConfig) setup(mod *WiFiModule) (err error) {
	if bruteforce.running.Load() {
		return errAlreadyRunning
	} else if err, bruteforce.target = mod.StringParam("wifi.bruteforce.target"); err != nil {
		return err
	} else if err, bruteforce.wordlist = mod.StringParam("wifi.bruteforce.wordlist"); err != nil {
		return err
	} else if err, bruteforce.workers = mod.IntParam("wifi.bruteforce.workers"); err != nil {
		return err
	} else if err, bruteforce.timeout = mod.IntParam("wifi.bruteforce.timeout"); err != nil {
		return err
	} else if err, bruteforce.wide = mod.BoolParam("wifi.bruteforce.wide"); err != nil {
		return err
	} else if err, bruteforce.stop_at_first = mod.BoolParam("wifi.bruteforce.stop_at_first"); err != nil {
		return err
	}

	// load targets
	bruteforce.targets = make([]string, 0)

	if bruteforce.target == "" {
		// all visible APs
		for _, ap := range mod.Session.WiFi.List() {
			if !ap.IsOpen() {
				target := ap.ESSID()
				if target == "<hidden>" || target == "" {
					target = ap.BSSID()
				}
				bruteforce.targets = append(bruteforce.targets, target)
			}
		}
	} else {
		bruteforce.targets = str.Comma(bruteforce.target)
	}

	nTargets := len(bruteforce.targets)
	if nTargets == 0 {
		return fmt.Errorf("no target selected with wifi.bruteforce.target='%s'", bruteforce.target)
	}

	mod.Info("selected %d target%s to bruteforce", nTargets, ops.Ternary(nTargets > 1, "s", ""))

	// load wordlist
	bruteforce.passwords = make([]string, 0)
	fp, err := os.Open(bruteforce.wordlist)
	if err != nil {
		return err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := str.Trim(scanner.Text())
		if line != "" {
			bruteforce.passwords = append(bruteforce.passwords, line)
		}
	}

	mod.Info("loaded %d passwords from %s", len(bruteforce.passwords), bruteforce.wordlist)

	mod.Info("starting %d workers ...", mod.bruteforce.workers)

	bruteforce.queue = async.NewQueue(mod.bruteforce.workers, mod.bruteforceWorker)

	bruteforce.running.Store(true)

	return
}

func (mod *WiFiModule) bruteforceWorker(arg async.Job) {
	job := arg.(bruteforceJob)
	defer job.done.Add(1)

	mod.Debug("got job %+v", job)

	if job.running.Load() {
		start := time.Now()

		if authenticated, err := wifiBruteforce(mod, job); err != nil {
			mod.Error("%v", err)
			// stop on error
			job.running.Store(false)
		} else if authenticated {
			// send event
			mod.Session.Events.Add("wifi.bruteforce.success", BruteforceSuccess{
				Elapsed:  time.Since(start),
				Iface:    job.iface,
				Target:   job.essid,
				Password: job.password,
			})
			if mod.bruteforce.stop_at_first {
				// stop if stop_at_first==true
				job.running.Store(false)
			}
		}
	}
}

func (mod *WiFiModule) showBruteforceProgress() {
	progress := 100.0 * (float64(mod.bruteforce.done.Load()) / float64(mod.bruteforce.todo))
	mod.State.Store("bruteforce.progress", progress)

	if mod.bruteforce.running.Load() {
		mod.Info("[%.2f%%] performed %d of %d bruteforcing attempts",
			progress,
			mod.bruteforce.done.Load(),
			mod.bruteforce.todo)
	}
}

func (mod *WiFiModule) startBruteforce() (err error) {
	var ifName string

	if mod.Running() {
		return errRecon
	} else if err = mod.bruteforce.setup(mod); err != nil {
		return err
	} else if err, ifName = mod.StringParam("wifi.interface"); err != nil {
		return err
	} else if ifName == "" {
		mod.iface = mod.Session.Interface
		ifName = mod.iface.Name()
	} else if mod.iface, err = network.FindInterface(ifName); err != nil {
		return fmt.Errorf("could not find interface %s: %v", ifName, err)
	} else if mod.iface == nil {
		return fmt.Errorf("could not find interface %s", ifName)
	}

	mod.Info("using interface %s", ifName)

	mod.bruteforce.todo = uint64(len(mod.bruteforce.passwords) * len(mod.bruteforce.targets))
	mod.bruteforce.done.Store(0)

	mod.Info("bruteforce running ...")

	go func() {
		go func() {
			if mod.bruteforce.wide {
				for _, password := range mod.bruteforce.passwords {
					for _, essid := range mod.bruteforce.targets {
						if mod.bruteforce.running.Load() {
							mod.bruteforce.queue.Add(async.Job(bruteforceJob{
								running:  &mod.bruteforce.running,
								done:     &mod.bruteforce.done,
								iface:    mod.iface.Name(),
								essid:    essid,
								password: password,
								timeout:  time.Second * time.Duration(mod.bruteforce.timeout),
							}))
						}
					}
				}
			} else {
				for _, essid := range mod.bruteforce.targets {
					for _, password := range mod.bruteforce.passwords {
						if mod.bruteforce.running.Load() {
							mod.bruteforce.queue.Add(async.Job(bruteforceJob{
								running:  &mod.bruteforce.running,
								done:     &mod.bruteforce.done,
								iface:    mod.iface.Name(),
								essid:    essid,
								password: password,
								timeout:  time.Second * time.Duration(mod.bruteforce.timeout),
							}))
						}
					}
				}
			}
		}()

		for mod.bruteforce.running.Load() && mod.bruteforce.done.Load() < mod.bruteforce.todo {
			time.Sleep(time.Second * time.Duration(mod.bruteforce.timeout))
			mod.showBruteforceProgress()
		}

		mod.bruteforce.running.Store(false)

		if mod.bruteforce.done.Load() == mod.bruteforce.todo {
			mod.Info("bruteforcing completed")
		} else {
			mod.Info("bruteforcing stopped")
		}
	}()

	return nil
}

func (mod *WiFiModule) stopBruteforce() error {
	if !mod.bruteforce.running.Load() {
		return errNotRunning
	}

	mod.Info("stopping bruteforcing ...")

	mod.bruteforce.running.Store(false)

	return nil
}
