package wifi

import (
	"errors"
	"net"
	"time"

	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/tui"
)

var errNoRecon = errors.New("Module wifi.ap requires module wifi.recon to be activated.")

func (mod *WiFiModule) parseApConfig() (err error) {
	var bssid string
	if err, mod.apConfig.SSID = mod.StringParam("wifi.ap.ssid"); err != nil {
		return
	} else if err, bssid = mod.StringParam("wifi.ap.bssid"); err != nil {
		return
	} else if mod.apConfig.BSSID, err = net.ParseMAC(network.NormalizeMac(bssid)); err != nil {
		return
	} else if err, mod.apConfig.Channel = mod.IntParam("wifi.ap.channel"); err != nil {
		return
	} else if err, mod.apConfig.Encryption = mod.BoolParam("wifi.ap.encryption"); err != nil {
		return
	}
	return
}

func (mod *WiFiModule) startAp() error {
	// we need channel hopping and packet injection for this
	if !mod.Running() {
		return errNoRecon
	} else if mod.apRunning {
		return session.ErrAlreadyStarted(mod.Name())
	}

	go func() {
		mod.apRunning = true
		defer func() {
			mod.apRunning = false
		}()

		enc := tui.Yellow("WPA2")
		if !mod.apConfig.Encryption {
			enc = tui.Green("Open")
		}
		mod.Info("sending beacons as SSID %s (%s) on channel %d (%s).",
			tui.Bold(mod.apConfig.SSID),
			mod.apConfig.BSSID.String(),
			mod.apConfig.Channel,
			enc)

		for seqn := uint16(0); mod.Running(); seqn++ {
			mod.writes.Add(1)
			defer mod.writes.Done()

			if err, pkt := packets.NewDot11Beacon(mod.apConfig, seqn); err != nil {
				mod.Error("could not create beacon packet: %s", err)
			} else {
				mod.injectPacket(pkt)
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()

	return nil
}
