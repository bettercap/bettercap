package net_sniff

import (
	"github.com/bettercap/bettercap/log"
	"time"
)

type SnifferStats struct {
	NumLocal    uint64
	NumMatched  uint64
	NumDumped   uint64
	NumWrote    uint64
	Started     time.Time
	FirstPacket time.Time
	LastPacket  time.Time
}

func NewSnifferStats() *SnifferStats {
	return &SnifferStats{
		NumLocal:    0,
		NumMatched:  0,
		NumDumped:   0,
		NumWrote:    0,
		Started:     time.Now(),
		FirstPacket: time.Time{},
		LastPacket:  time.Time{},
	}
}

func (s *SnifferStats) Print() error {
	first := "never"
	last := "never"

	if !s.FirstPacket.IsZero() {
		first = s.FirstPacket.String()
	}
	if !s.LastPacket.IsZero() {
		last = s.LastPacket.String()
	}

	log.Info("Sniffer Started    : %s", s.Started)
	log.Info("First Packet Seen  : %s", first)
	log.Info("Last Packet Seen   : %s", last)
	log.Info("Local Packets      : %d", s.NumLocal)
	log.Info("Matched Packets    : %d", s.NumMatched)
	log.Info("Dumped Packets     : %d", s.NumDumped)
	log.Info("Wrote Packets      : %d", s.NumWrote)

	return nil
}
