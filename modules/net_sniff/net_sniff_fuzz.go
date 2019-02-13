package net_sniff

import (
	"math/rand"
	"strings"

	"github.com/google/gopacket"

	"github.com/evilsocket/islazy/str"
)

var mutators = []func(byte) byte{
	func(b byte) byte {
		return byte(rand.Intn(256) & 0xff)
	},
	func(b byte) byte {
		return byte(b << uint(rand.Intn(9)))
	},
	func(b byte) byte {
		return byte(b >> uint(rand.Intn(9)))
	},
}

func (mod *Sniffer) fuzz(data []byte) int {
	changes := 0
	for off, b := range data {
		if rand.Float64() > mod.fuzzRatio {
			continue
		}

		data[off] = mutators[rand.Intn(len(mutators))](b)
		changes++
	}
	return changes
}

func (mod *Sniffer) doFuzzing(pkt gopacket.Packet) {
	if rand.Float64() > mod.fuzzRate {
		return
	}

	layersChanged := 0
	bytesChanged := 0

	for _, fuzzLayerType := range mod.fuzzLayers {
		for _, layer := range pkt.Layers() {
			if layer.LayerType().String() == fuzzLayerType {
				fuzzData := layer.LayerContents()
				changes := mod.fuzz(fuzzData)
				if changes > 0 {
					layersChanged++
					bytesChanged += changes
				}

			}
		}
	}

	if bytesChanged > 0 {
		logFn := mod.Info
		if mod.fuzzSilent {
			logFn = mod.Debug
		}
		logFn("changed %d bytes in %d layers.", bytesChanged, layersChanged)
		if err := mod.Session.Queue.Send(pkt.Data()); err != nil {
			mod.Error("error sending fuzzed packet: %s", err)
		}
	}
}

func (mod *Sniffer) configureFuzzing() (err error) {
	layers := ""

	if err, layers = mod.StringParam("net.fuzz.layers"); err != nil {
		return
	} else {
		mod.fuzzLayers = str.Comma(layers)
	}

	if err, mod.fuzzRate = mod.DecParam("net.fuzz.rate"); err != nil {
		return
	}

	if err, mod.fuzzRatio = mod.DecParam("net.fuzz.ratio"); err != nil {
		return
	}

	if err, mod.fuzzSilent = mod.BoolParam("net.fuzz.silent"); err != nil {
		return
	}

	return
}

func (mod *Sniffer) StartFuzzing() error {
	if mod.fuzzActive {
		return nil
	}

	if err := mod.configureFuzzing(); err != nil {
		return err
	} else if !mod.Running() {
		if err := mod.Start(); err != nil {
			return err
		}
	}

	mod.fuzzActive = true

	mod.Info("active on layer types %s (rate:%f ratio:%f)", strings.Join(mod.fuzzLayers, ","), mod.fuzzRate, mod.fuzzRatio)

	return nil
}

func (mod *Sniffer) StopFuzzing() error {
	mod.fuzzActive = false
	return nil
}
