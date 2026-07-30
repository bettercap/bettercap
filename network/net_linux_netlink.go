package network

import (
	"fmt"
	"net"
	"sync"

	"github.com/mdlayher/genetlink"
	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

// SetInterfaceChannel's old path forks a whole new `iw` process on every
// single channel hop -- process creation, ELF/dynamic-linker work to load
// `iw` and its own libraries, then finally one short netlink message sent
// and the process torn down again -- to do something that's really just
// "send one short message over a socket". Since bettercap is already a
// running process, it can open that same kind of socket itself and speak
// the identical nl80211 protocol `iw` does, with no process spawn at all.
//
// One connection + one resolved nl80211 family ID, opened lazily on first
// use and kept for the life of the process, instead of paying full
// process-spawn cost on every hop (confirmed a real, measured cost on the
// Mi Mix 3 -- see pwndroid notes/02-monitor-mode.md). Falls back to the
// process-based iw/iwconfig path in SetInterfaceChannel on any failure
// here, so this can only ever be as slow as before, never worse (one
// fast, already-failing netlink call added on top of the existing
// fallback, not instead of it).
var (
	nlOnce    sync.Once
	nlConn    *genetlink.Conn
	nlFamily  genetlink.Family
	nlInitErr error
	// mdlayher/netlink's Conn isn't documented as safe for concurrent
	// Execute() calls from multiple goroutines, and not every caller in
	// this package goes through wifi_hopping.go's own chanLock (see
	// GetSupportedFrequencies, still iw-based) -- serialize here instead
	// of assuming every future caller will remember to.
	nlMu sync.Mutex
)

func nlInit() {
	nlOnce.Do(func() {
		conn, err := genetlink.Dial(nil)
		if err != nil {
			nlInitErr = fmt.Errorf("netlink: dial: %w", err)
			return
		}
		family, err := conn.GetFamily("nl80211")
		if err != nil {
			conn.Close()
			nlInitErr = fmt.Errorf("netlink: resolve nl80211 family: %w", err)
			return
		}
		nlConn = conn
		nlFamily = family
	})
}

// Direct nl80211 equivalent of `iw dev <iface> set freq <freqMHz>` -- same
// NL80211_CMD_SET_CHANNEL command, same attributes as what `iw` itself
// sends for a device-level (not wiphy-level) channel set, over the shared
// persistent connection above instead of a fresh process. Plain 20MHz,
// no-HT width (NL80211_CHAN_NO_HT) -- matches what a monitor-mode
// capture/injection interface actually needs, no channel bonding.
func setInterfaceChannelNetlink(iface string, freqMHz int) error {
	nlInit()
	if nlInitErr != nil {
		return nlInitErr
	}

	ifi, err := net.InterfaceByName(iface)
	if err != nil {
		return fmt.Errorf("netlink: interface %s: %w", iface, err)
	}

	ae := netlink.NewAttributeEncoder()
	ae.Uint32(unix.NL80211_ATTR_IFINDEX, uint32(ifi.Index))
	ae.Uint32(unix.NL80211_ATTR_WIPHY_FREQ, uint32(freqMHz))
	ae.Uint32(unix.NL80211_ATTR_WIPHY_CHANNEL_TYPE, unix.NL80211_CHAN_NO_HT)
	data, err := ae.Encode()
	if err != nil {
		return fmt.Errorf("netlink: encode attributes: %w", err)
	}

	req := genetlink.Message{
		Header: genetlink.Header{
			Command: unix.NL80211_CMD_SET_CHANNEL,
			Version: nlFamily.Version,
		},
		Data: data,
	}

	nlMu.Lock()
	defer nlMu.Unlock()
	if _, err := nlConn.Execute(req, nlFamily.ID, netlink.Request|netlink.Acknowledge); err != nil {
		return fmt.Errorf("netlink: NL80211_CMD_SET_CHANNEL: %w", err)
	}
	return nil
}
