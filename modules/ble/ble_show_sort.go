// +build !windows

package ble

import (
	"github.com/bettercap/bettercap/network"
)

type ByBLERSSISorter []*network.BLEDevice

func (a ByBLERSSISorter) Len() int      { return len(a) }
func (a ByBLERSSISorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByBLERSSISorter) Less(i, j int) bool {
	if a[i].RSSI == a[j].RSSI {
		return a[i].Device.ID() < a[j].Device.ID()
	}
	return a[i].RSSI > a[j].RSSI
}

type ByBLEMacSorter []*network.BLEDevice

func (a ByBLEMacSorter) Len() int      { return len(a) }
func (a ByBLEMacSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByBLEMacSorter) Less(i, j int) bool {
	return a[i].Device.ID() < a[j].Device.ID()
}

type ByBLESeenSorter []*network.BLEDevice

func (a ByBLESeenSorter) Len() int           { return len(a) }
func (a ByBLESeenSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByBLESeenSorter) Less(i, j int) bool { return a[i].LastSeen.Before(a[j].LastSeen) }
