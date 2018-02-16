package modules

import (
	"net"
	"sync"
)

type WiFiStationStats struct {
	Bytes uint64
}

type WiFiStats struct {
	sync.Mutex
	stats map[string]*WiFiStationStats
}

func NewWiFiStats() *WiFiStats {
	return &WiFiStats{
		stats: make(map[string]*WiFiStationStats),
	}
}

func (s *WiFiStats) Collect(station net.HardwareAddr, bytes uint64) {
	s.Lock()
	defer s.Unlock()

	bssid := station.String()
	if sstats, found := s.stats[bssid]; found == true {
		sstats.Bytes += bytes
	} else {
		s.stats[bssid] = &WiFiStationStats{Bytes: bytes}
	}
}

func (s *WiFiStats) For(station net.HardwareAddr) uint64 {
	s.Lock()
	defer s.Unlock()

	bssid := station.String()
	if sstats, found := s.stats[bssid]; found == true {
		return sstats.Bytes
	}
	return 0
}
