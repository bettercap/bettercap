package network

import (
	"sync"
)

const NO_CHANNEL = -1

var (
	currChannels    = make(map[string]int)
	currChannelLock = sync.Mutex{}
)

func GetInterfaceChannel(iface string) int {
	currChannelLock.Lock()
	defer currChannelLock.Unlock()
	if curr, found := currChannels[iface]; found {
		return curr
	}
	return NO_CHANNEL
}

func SetInterfaceCurrentChannel(iface string, channel int) {
	currChannelLock.Lock()
	defer currChannelLock.Unlock()
	currChannels[iface] = channel
}
