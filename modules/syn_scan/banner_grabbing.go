package syn_scan

import (
	"fmt"
	"github.com/bettercap/bettercap/network"

	"github.com/evilsocket/islazy/async"
)

type bannerGrabberFn func(mod *SynScanner, ip string, port int) string

type grabberJob struct {
	Host *network.Endpoint
	Port *OpenPort
}

func (mod *SynScanner) bannerGrabber(arg async.Job) {
	job := arg.(grabberJob)
	if job.Port.Proto != "tcp" {
		return
	}

	ip := job.Host.IpAddress
	port := job.Port.Port
	sport := fmt.Sprintf("%d", port)

	fn := tcpGrabber
	if port == 80 || port == 443 || sport[0] == '8' {
		fn = httpGrabber
	}

	mod.Debug("grabbing banner for %s:%d", ip, port)
	job.Port.Banner = fn(mod, ip, port)
	if job.Port.Banner != "" {
		mod.Info("found banner for %s:%d -> %s", ip, port, job.Port.Banner)
	}
}
