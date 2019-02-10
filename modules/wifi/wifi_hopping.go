package wifi

import (
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
)

func (w *WiFiModule) onChannel(channel int, cb func()) {
	w.chanLock.Lock()
	defer w.chanLock.Unlock()

	prev := w.stickChan
	w.stickChan = channel

	if err := network.SetInterfaceChannel(w.Session.Interface.Name(), channel); err != nil {
		log.Warning("error while hopping to channel %d: %s", channel, err)
	} else {
		log.Debug("hopped on channel %d", channel)
	}

	cb()

	w.stickChan = prev
}

func (w *WiFiModule) channelHopper() {
	w.reads.Add(1)
	defer w.reads.Done()

	log.Info("channel hopper started.")

	for w.Running() {
		delay := w.hopPeriod
		// if we have both 2.4 and 5ghz capabilities, we have
		// more channels, therefore we need to increase the time
		// we hop on each one otherwise me lose information
		if len(w.frequencies) > 14 {
			delay = delay * 2
		}

		frequencies := w.frequencies

	loopCurrentChannels:
		for _, frequency := range frequencies {
			channel := network.Dot11Freq2Chan(frequency)
			// stick to the access point channel as long as it's selected
			// or as long as we're deauthing on it
			if w.stickChan != 0 {
				channel = w.stickChan
			}

			log.Debug("hopping on channel %d", channel)

			w.chanLock.Lock()
			if err := network.SetInterfaceChannel(w.Session.Interface.Name(), channel); err != nil {
				log.Warning("error while hopping to channel %d: %s", channel, err)
			}
			w.chanLock.Unlock()

			select {
			case _ = <-w.hopChanges:
				log.Debug("hop changed")
				break loopCurrentChannels
			case <-time.After(delay):
				if !w.Running() {
					return
				}
			}
		}
	}
}
