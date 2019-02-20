package hid

import (
	"github.com/bettercap/bettercap/network"
)

type ByHIDMacSorter []*network.HIDDevice

func (a ByHIDMacSorter) Len() int      { return len(a) }
func (a ByHIDMacSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByHIDMacSorter) Less(i, j int) bool {
	return a[i].Address < a[j].Address
}

type ByHIDSeenSorter []*network.HIDDevice

func (a ByHIDSeenSorter) Len() int           { return len(a) }
func (a ByHIDSeenSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByHIDSeenSorter) Less(i, j int) bool { return a[i].LastSeen.Before(a[j].LastSeen) }
