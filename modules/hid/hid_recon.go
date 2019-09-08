package hid

import (
	"time"

	"github.com/bettercap/nrf24"
	"github.com/google/gousb"
)

func (mod *HIDRecon) doHopping() {
	mod.writeLock.Lock()
	defer mod.writeLock.Unlock()

	if mod.inPromMode == false {
		if err := mod.dongle.EnterPromiscMode(); err != nil {
			mod.Error("error entering promiscuous mode: %v", err)
		} else {
			mod.inSniffMode = false
			mod.inPromMode = true
			mod.Debug("device entered promiscuous mode")
		}
	}

	if time.Since(mod.lastHop) >= mod.hopPeriod {
		mod.channel++
		if mod.channel > nrf24.TopChannel {
			mod.channel = 1
		}
		if err := mod.dongle.SetChannel(mod.channel); err != nil {
			if err == gousb.ErrorNoDevice || err == gousb.TransferStall {
				mod.Error("device disconnected, stopping module")
				mod.forceStop()
				return
			} else {
				mod.Warning("error hopping on channel %d: %v", mod.channel, err)
			}
		} else {
			mod.lastHop = time.Now()
		}
	}
}

func (mod *HIDRecon) onDeviceDetected(buf []byte) {
	if sz := len(buf); sz >= 5 {
		addr, payload := buf[0:5], buf[5:]
		mod.Debug("detected device %x on channel %d (payload:%x)\n", addr, mod.channel, payload)
		if isNew, dev := mod.Session.HID.AddIfNew(addr, mod.channel, payload); isNew {
			// sniff for a while in order to detect the device type
			go func() {
				prevSilent := mod.sniffSilent

				if err := mod.setSniffMode(dev.Address, true); err == nil {
					mod.Debug("detecting device type ...")
					defer func() {
						mod.sniffLock.Unlock()
						mod.setSniffMode("clear", prevSilent)
					}()
					// make sure nobody can sniff to another
					// address until we're not done here...
					mod.sniffLock.Lock()

					time.Sleep(mod.sniffPeriod)
				} else {
					mod.Warning("error while sniffing %s: %v", dev.Address, err)
				}
			}()
		}
	}
}

func (mod *HIDRecon) devPruner() {
	mod.waitGroup.Add(1)
	defer mod.waitGroup.Done()

	maxDeviceTTL := time.Duration(mod.devTTL) * time.Second
	mod.Debug("devices pruner started with ttl %v", maxDeviceTTL)
	for mod.Running() {
		for _, dev := range mod.Session.HID.Devices() {
			sinceLastSeen := time.Since(dev.LastSeen)
			if sinceLastSeen > maxDeviceTTL {
				mod.Debug("device %s not seen in %s, removing.", dev.Address, sinceLastSeen)
				mod.Session.HID.Remove(dev.Address)
			}
		}
		time.Sleep(30 * time.Second)
	}
}

func (mod *HIDRecon) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.waitGroup.Add(1)
		defer mod.waitGroup.Done()

		go mod.devPruner()

		mod.Info("hopping on %d channels every %s", nrf24.TopChannel, mod.hopPeriod)
		for mod.Running() {
			if mod.isSniffing() {
				mod.doPing()
			} else {
				mod.doHopping()
			}

			if mod.isInjecting() {
				mod.doInjection()
				mod.setInjectionMode("clear")
				continue
			}

			buf, err := mod.dongle.ReceivePayload()
			if err != nil {
				if err == gousb.ErrorNoDevice || err == gousb.TransferStall {
					mod.Error("device disconnected, stopping module")
					mod.forceStop()
					return
				}
				mod.Warning("error receiving payload from channel %d: %v", mod.channel, err)
				continue
			}

			if mod.isSniffing() {
				mod.onSniffedBuffer(buf)
			} else {
				mod.onDeviceDetected(buf)
			}
		}

		mod.Debug("stopped")
	})
}
