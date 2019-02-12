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

func (s *Sniffer) fuzz(data []byte) int {
	changes := 0
	for off, b := range data {
		if rand.Float64() > s.fuzzRatio {
			continue
		}

		data[off] = mutators[rand.Intn(len(mutators))](b)
		changes++
	}
	return changes
}

func (s *Sniffer) doFuzzing(pkt gopacket.Packet) {
	if rand.Float64() > s.fuzzRate {
		return
	}

	layersChanged := 0
	bytesChanged := 0

	for _, fuzzLayerType := range s.fuzzLayers {
		for _, layer := range pkt.Layers() {
			if layer.LayerType().String() == fuzzLayerType {
				fuzzData := layer.LayerContents()
				changes := s.fuzz(fuzzData)
				if changes > 0 {
					layersChanged++
					bytesChanged += changes
				}

			}
		}
	}

	if bytesChanged > 0 {
		logFn := s.Info
		if s.fuzzSilent {
			logFn = s.Debug
		}
		logFn("changed %d bytes in %d layers.", bytesChanged, layersChanged)
		if err := s.Session.Queue.Send(pkt.Data()); err != nil {
			s.Error("error sending fuzzed packet: %s", err)
		}
	}
}

func (s *Sniffer) configureFuzzing() (err error) {
	layers := ""

	if err, layers = s.StringParam("net.fuzz.layers"); err != nil {
		return
	} else {
		s.fuzzLayers = str.Comma(layers)
	}

	if err, s.fuzzRate = s.DecParam("net.fuzz.rate"); err != nil {
		return
	}

	if err, s.fuzzRatio = s.DecParam("net.fuzz.ratio"); err != nil {
		return
	}

	if err, s.fuzzSilent = s.BoolParam("net.fuzz.silent"); err != nil {
		return
	}

	return
}

func (s *Sniffer) StartFuzzing() error {
	if s.fuzzActive {
		return nil
	}

	if err := s.configureFuzzing(); err != nil {
		return err
	} else if !s.Running() {
		if err := s.Start(); err != nil {
			return err
		}
	}

	s.fuzzActive = true

	s.Info("active on layer types %s (rate:%f ratio:%f)", strings.Join(s.fuzzLayers, ","), s.fuzzRate, s.fuzzRatio)

	return nil
}

func (s *Sniffer) StopFuzzing() error {
	s.fuzzActive = false
	return nil
}
