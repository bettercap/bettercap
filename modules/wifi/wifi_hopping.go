package wifi

import (
	"net"
	"time"

	"github.com/bettercap/bettercap/network"
)

func (mod *WiFiModule) isInterfaceConnected() bool {
	ifaces, err := net.Interfaces()
	if err != nil {
		mod.Error("error while enumerating interfaces: %s", err)
		return false
	}

	for _, iface := range ifaces {
		if mod.iface.HwAddress == network.NormalizeMac(iface.HardwareAddr.String()) {
			return true
		}
	}

	return false
}

func (mod *WiFiModule) hopUnlocked(channel int) (mustStop bool) {
	// mod.Debug("hopping on channel %d", channel)

	if err := network.SetInterfaceChannel(mod.iface.Name(), channel); err != nil {
		// check if the device has been disconnected
		if mod.isInterfaceConnected() == false {
			mod.Error("interface %s disconnected, stopping module", mod.iface.Name())
			mustStop = true
		} else {
			mod.Warning("error while hopping to channel %d: %s", channel, err)
		}
	}

	return
}

func (mod *WiFiModule) hop(channel int) (mustStop bool) {
	mod.chanLock.Lock()
	defer mod.chanLock.Unlock()

	return mod.hopUnlocked(channel)
}

func (mod *WiFiModule) onChannel(channel int, cb func()) {
	mod.chanLock.Lock()
	defer mod.chanLock.Unlock()

	prev := mod.stickChan
	mod.stickChan = channel

	mod.hopUnlocked(channel)

	cb()

	mod.stickChan = prev
}

func (mod *WiFiModule) channelHopper() {
	mod.reads.Add(1)
	defer mod.reads.Done()

	mod.Info("channel hopper started.")

	for mod.Running() {
		delay := mod.hopPeriod
		// if we have both 2.4 and 5ghz capabilities, we have
		// more channels, therefore we need to increase the time
		// we hop on each one otherwise me lose information
		if len(mod.frequencies) > 14 {
			delay = delay * 2
		}

		frequencies := mod.frequencies

	loopCurrentChannels:
		for _, frequency := range frequencies {
			channel := network.Dot11Freq2Chan(frequency)
			// stick to the access point channel as long as it's selected
			// or as long as we're deauthing on it
			if mod.stickChan != 0 {
				channel = mod.stickChan
			}

			if stop := mod.hop(channel); stop {
				mod.forcedStop()
				return
			}

			select {
			case <-mod.hopChanges:
				mod.Debug("hop changed")
				break loopCurrentChannels
			case <-time.After(delay):
				if !mod.Running() {
					return
				}
			}
		}
	}
}
