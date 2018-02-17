package modules

import (
	"net"
	"sync"
)

type WiFiStationStats struct {
	Sent     uint64
	Received uint64
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

func (s *WiFiStats) CollectSent(station net.HardwareAddr, bytes uint64) {
	s.Lock()
	defer s.Unlock()

	bssid := station.String()
	if sstats, found := s.stats[bssid]; found == true {
		sstats.Sent += bytes
	} else {
		s.stats[bssid] = &WiFiStationStats{Sent: bytes}
	}
}

func (s *WiFiStats) CollectReceived(station net.HardwareAddr, bytes uint64) {
	s.Lock()
	defer s.Unlock()

	bssid := station.String()
	if sstats, found := s.stats[bssid]; found == true {
		sstats.Received += bytes
	} else {
		s.stats[bssid] = &WiFiStationStats{Received: bytes}
	}
}

func (s *WiFiStats) SentFrom(station net.HardwareAddr) uint64 {
	s.Lock()
	defer s.Unlock()

	bssid := station.String()
	if sstats, found := s.stats[bssid]; found == true {
		return sstats.Sent
	}
	return 0
}

func (s *WiFiStats) SentTo(station net.HardwareAddr) uint64 {
	s.Lock()
	defer s.Unlock()

	bssid := station.String()
	if sstats, found := s.stats[bssid]; found == true {
		return sstats.Received
	}
	return 0
}
