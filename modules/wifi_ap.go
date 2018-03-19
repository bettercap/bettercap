package modules

import (
	"errors"
	"net"
	"time"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"
)

var errNoRecon = errors.New("Module wifi.ap requires module wifi.recon to be activated.")

func (w *WiFiModule) parseApConfig() (err error) {
	var bssid string

	if err, w.apConfig.SSID = w.StringParam("wifi.ap.ssid"); err != nil {
		return
	} else if err, bssid = w.StringParam("wifi.ap.bssid"); err != nil {
		return
	} else if w.apConfig.BSSID, err = net.ParseMAC(network.NormalizeMac(bssid)); err != nil {
		return
	} else if err, w.apConfig.Channel = w.IntParam("wifi.ap.channel"); err != nil {
		return
	} else if err, w.apConfig.Encryption = w.BoolParam("wifi.ap.encryption"); err != nil {
		return
	}

	return
}

func (w *WiFiModule) startAp() error {
	// we need channel hopping and packet injection for this
	if w.Running() == false {
		return errNoRecon
	} else if w.apRunning {
		return session.ErrAlreadyStarted
	}

	go func() {
		w.apRunning = true
		defer func() {
			w.apRunning = false
		}()

		enc := core.Yellow("WPA2")
		if w.apConfig.Encryption == false {
			enc = core.Green("Open")
		}
		log.Info("Sending beacons as SSID %s (%s) on channel %d (%s).",
			core.Bold(w.apConfig.SSID),
			w.apConfig.BSSID.String(),
			w.apConfig.Channel,
			enc)

		for seqn := uint16(0); w.Running(); seqn++ {
			w.writes.Add(1)
			defer w.writes.Done()

			if err, pkt := packets.NewDot11Beacon(w.apConfig, seqn); err != nil {
				log.Error("Could not create beacon packet: %s", err)
			} else {
				w.injectPacket(pkt)
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()

	return nil
}
