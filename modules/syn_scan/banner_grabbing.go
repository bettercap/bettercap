package syn_scan

import (
	"fmt"
	"time"

	"github.com/evilsocket/islazy/async"
)

const bannerGrabTimeout = time.Duration(5) * time.Second

type bannerGrabberFn func(mod *SynScanner, ip string, port int) string

type grabberJob struct {
	IP   string
	Port *OpenPort
}

func (mod *SynScanner) bannerGrabber(arg async.Job) {
	job := arg.(grabberJob)
	if job.Port.Proto != "tcp" {
		return
	}

	ip := job.IP
	port := job.Port.Port
	sport := fmt.Sprintf("%d", port)

	fn := tcpGrabber
	if port == 80 || port == 443 || sport[0] == '8' {
		fn = httpGrabber
	} else if port == 53 || port == 5353 {
		fn = dnsGrabber
	}

	mod.Debug("grabbing banner for %s:%d", ip, port)
	job.Port.Banner = fn(mod, ip, port)
	if job.Port.Banner != "" {
		mod.Info("found banner for %s:%d -> %s", ip, port, job.Port.Banner)
	}
}
