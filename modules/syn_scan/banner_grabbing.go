package syn_scan

import (
	"github.com/bettercap/bettercap/network"

	"github.com/evilsocket/islazy/async"
)

type bannerGrabberFn func(mod *SynScanner, ip string, port int) string

type grabberJob struct {
	Host *network.Endpoint
	Port *OpenPort
}

var tcpBannerGrabbers = map[int]bannerGrabberFn{
	80:   httpGrabber,
	8080: httpGrabber,
	443:  httpGrabber,
	8443: httpGrabber,
}

func (mod *SynScanner) bannerGrabber(arg async.Job) {
	job := arg.(grabberJob)
	if job.Port.Proto != "tcp" {
		return
	}

	ip := job.Host.IpAddress
	port := job.Port.Port
	fn, found := tcpBannerGrabbers[port]
	if !found {
		fn = tcpGrabber
	}

	mod.Debug("grabbing banner for %s:%d", ip, port)
	job.Port.Banner = fn(mod, ip, port)
	if job.Port.Banner != "" {
		mod.Info("found banner for %s:%d -> %s", ip, port, job.Port.Banner)
	}
}
