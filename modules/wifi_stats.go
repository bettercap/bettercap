package modules

import (
	"net"
	"strings"
	"sync"
)

type WiFiStationStats struct {
	Sent       uint64
	Received   uint64
	Encryption map[string]bool
}

func NewWiFiStationStats(sent uint64, recvd uint64) *WiFiStationStats {
	return &WiFiStationStats{
		Sent:       sent,
		Received:   recvd,
		Encryption: make(map[string]bool),
	}
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
		s.stats[bssid] = NewWiFiStationStats(bytes, 0)
	}
}

func (s *WiFiStats) CollectReceived(station net.HardwareAddr, bytes uint64) {
	s.Lock()
	defer s.Unlock()

	bssid := station.String()
	if sstats, found := s.stats[bssid]; found == true {
		sstats.Received += bytes
	} else {
		s.stats[bssid] = NewWiFiStationStats(0, bytes)
	}
}

func (s *WiFiStats) ResetEncryption(station net.HardwareAddr) {
	s.Lock()
	defer s.Unlock()

	bssid := station.String()
	if sstats, found := s.stats[bssid]; found == true {
		sstats.Encryption = make(map[string]bool)
	}
}

func (s *WiFiStats) CollectEncryption(station net.HardwareAddr, enc string) {
	s.Lock()
	defer s.Unlock()

	bssid := station.String()
	if sstats, found := s.stats[bssid]; found == true {
		sstats.Encryption[enc] = true
	} else {
		stats := NewWiFiStationStats(0, 0)
		stats.Encryption[enc] = true
		s.stats[bssid] = stats
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

func (s *WiFiStats) EncryptionOf(station net.HardwareAddr) string {
	s.Lock()
	defer s.Unlock()

	bssid := station.String()
	if sstats, found := s.stats[bssid]; found == true {
		unique := make([]string, 0)
		for key := range sstats.Encryption {
			unique = append(unique, key)
		}
		return strings.Join(unique, ", ")
	}
	return ""
}
