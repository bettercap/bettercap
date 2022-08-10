package network

import (
	"fmt"
	"time"

	"github.com/evilsocket/islazy/tui"
	"github.com/google/gopacket/pcap"
)

const (
	PCAP_DEFAULT_SETRF   = false
	PCAP_DEFAULT_SNAPLEN = 65536
	PCAP_DEFAULT_BUFSIZE = 2_097_152
	PCAP_DEFAULT_PROMISC = true
	PCAP_DEFAULT_TIMEOUT = pcap.BlockForever
)

var CAPTURE_DEFAULTS = CaptureOptions{
	Monitor: PCAP_DEFAULT_SETRF,
	Snaplen: PCAP_DEFAULT_SNAPLEN,
	Bufsize: PCAP_DEFAULT_BUFSIZE,
	Promisc: PCAP_DEFAULT_PROMISC,
	Timeout: PCAP_DEFAULT_TIMEOUT,
}

type CaptureOptions struct {
	Monitor bool
	Snaplen int
	Bufsize int
	Promisc bool
	Timeout time.Duration
}

func CaptureWithOptions(ifName string, options CaptureOptions) (*pcap.Handle, error) {
	Debug("creating capture for '%s' with options: %+v", ifName, options)

	ihandle, err := pcap.NewInactiveHandle(ifName)
	if err != nil {
		return nil, fmt.Errorf("error while opening interface %s: %s", ifName, err)
	}
	defer ihandle.CleanUp()

	if options.Monitor {
		if err = ihandle.SetRFMon(true); err != nil {
			return nil, fmt.Errorf("error while setting interface %s in monitor mode: %s", tui.Bold(ifName), err)
		}
	}

	if err = ihandle.SetSnapLen(options.Snaplen); err != nil {
		return nil, fmt.Errorf("error while settng snapshot length: %s", err)
	} else if err = ihandle.SetBufferSize(options.Bufsize); err != nil {
		return nil, fmt.Errorf("error while settng buffer size: %s", err)
	} else if err = ihandle.SetPromisc(options.Promisc); err != nil {
		return nil, fmt.Errorf("error while settng promiscuous mode to %v: %s", options.Promisc, err)
	} else if err = ihandle.SetTimeout(options.Timeout); err != nil {
		return nil, fmt.Errorf("error while settng snapshot length: %s", err)
	}

	return ihandle.Activate()
}

func Capture(ifName string) (*pcap.Handle, error) {
	return CaptureWithOptions(ifName, CAPTURE_DEFAULTS)
}

func CaptureWithTimeout(ifName string, timeout time.Duration) (*pcap.Handle, error) {
	var opts = CAPTURE_DEFAULTS
	opts.Timeout = timeout
	return CaptureWithOptions(ifName, opts)
}
