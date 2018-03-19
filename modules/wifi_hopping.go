package modules

import (
	"time"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/network"
)

func (w *WiFiModule) onChannel(channel int, cb func()) {
	prev := w.stickChan
	w.stickChan = channel

	if err := network.SetInterfaceChannel(w.Session.Interface.Name(), channel); err != nil {
		log.Warning("Error while hopping to channel %d: %s", channel, err)
	} else {
		log.Debug("Hopped on channel %d", channel)
	}

	cb()

	w.stickChan = prev
}

func (w *WiFiModule) channelHopper() {
	w.reads.Add(1)
	defer w.reads.Done()

	log.Info("Channel hopper started.")

	for w.Running() == true {
		delay := w.hopPeriod
		// if we have both 2.4 and 5ghz capabilities, we have
		// more channels, therefore we need to increase the time
		// we hop on each one otherwise me lose information
		if len(w.frequencies) > 14 {
			delay = delay * 2
		}

		frequencies := w.frequencies
		for _, frequency := range frequencies {
			channel := network.Dot11Freq2Chan(frequency)
			// stick to the access point channel as long as it's selected
			// or as long as we're deauthing on it
			if w.stickChan != 0 {
				channel = w.stickChan
			}

			log.Debug("Hopping on channel %d", channel)

			if err := network.SetInterfaceChannel(w.Session.Interface.Name(), channel); err != nil {
				log.Warning("Error while hopping to channel %d: %s", channel, err)
			}

			time.Sleep(delay)
			if w.Running() == false {
				return
			}
		}
	}
}
