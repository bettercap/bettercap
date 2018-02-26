package network

import (
	"time"
)

type BLEDevice struct {
	LastSeen time.Time
}

func NewBLEDevice() *BLEDevice {
	return &BLEDevice{
		LastSeen: time.Now(),
	}
}
