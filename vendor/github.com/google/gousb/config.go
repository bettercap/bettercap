// Copyright 2013 Google Inc.  All rights reserved.
// Copyright 2016 the gousb Authors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gousb

import (
	"fmt"
	"sync"
)

// ConfigDesc contains the information about a USB device configuration,
// extracted from the device descriptor.
type ConfigDesc struct {
	// Number is the configuration number.
	Number int
	// SelfPowered is true if the device is powered externally, i.e. not
	// drawing power from the USB bus.
	SelfPowered bool
	// RemoteWakeup is true if the device supports remote wakeup, i.e.
	// an external signal that will wake up a suspended USB device. An example
	// might be a keyboard that can wake up through a keypress after
	// the host put it in suspend mode. Note that gousb does not support
	// device power management, RemoteWakeup only refers to the reported device
	// capability.
	RemoteWakeup bool
	// MaxPower is the maximum current the device draws from the USB bus
	// in this configuration.
	MaxPower Milliamperes
	// Interfaces has a list of USB interfaces available in this configuration.
	Interfaces []InterfaceDesc

	iConfiguration int // index of a string descriptor describing this configuration
}

// String returns the human-readable description of the configuration descriptor.
func (c ConfigDesc) String() string {
	return fmt.Sprintf("Configuration %d", c.Number)
}

func (c ConfigDesc) intfDesc(num, alt int) (*InterfaceSetting, error) {
	if num < 0 || num >= len(c.Interfaces) {
		return nil, fmt.Errorf("interface %d not found, available interfaces 0..%d", num, len(c.Interfaces)-1)
	}
	ifInfo := c.Interfaces[num]
	if alt < 0 || alt >= len(ifInfo.AltSettings) {
		return nil, fmt.Errorf("alternate setting %d not found for %s, available alt settings 0..%d", alt, ifInfo, len(ifInfo.AltSettings)-1)
	}
	return &ifInfo.AltSettings[alt], nil
}

// Config represents a USB device set to use a particular configuration.
// Only one Config of a particular device can be used at any one time.
// To access device endpoints, claim an interface and it's alternate
// setting number through a call to Interface().
type Config struct {
	Desc ConfigDesc

	dev *Device

	// Claimed interfaces
	mu      sync.Mutex
	claimed map[int]bool
}

// Close releases the underlying device, allowing the caller to switch the device to a different configuration.
func (c *Config) Close() error {
	if c.dev == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.claimed) > 0 {
		var ifs []int
		for k := range c.claimed {
			ifs = append(ifs, k)
		}
		return fmt.Errorf("failed to release %s, interfaces %v are still open", c, ifs)
	}
	c.dev.mu.Lock()
	defer c.dev.mu.Unlock()
	c.dev.claimed = nil
	c.dev = nil
	return nil
}

// String returns the human-readable description of the configuration.
func (c *Config) String() string {
	return fmt.Sprintf("%s,config=%d", c.dev.String(), c.Desc.Number)
}

// Interface claims and returns an interface on a USB device.
// num specifies the number of an interface to claim, and alt specifies the
// alternate setting number for that interface.
func (c *Config) Interface(num, alt int) (*Interface, error) {
	if c.dev == nil {
		return nil, fmt.Errorf("Interface(%d, %d) called on %s after Close", num, alt, c)
	}

	altInfo, err := c.Desc.intfDesc(num, alt)
	if err != nil {
		return nil, fmt.Errorf("descriptor of interface (%d, %d) in %s: %v", num, alt, c, err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.claimed[num] {
		return nil, fmt.Errorf("interface %d on %s is already claimed", num, c)
	}

	// Claim the interface
	if err := c.dev.ctx.libusb.claim(c.dev.handle, uint8(num)); err != nil {
		return nil, fmt.Errorf("failed to claim interface %d on %s: %v", num, c, err)
	}

	// Select an alternate setting if needed (device has multiple alternate settings).
	if len(c.Desc.Interfaces[num].AltSettings) > 1 {
		if err := c.dev.ctx.libusb.setAlt(c.dev.handle, uint8(num), uint8(alt)); err != nil {
			c.dev.ctx.libusb.release(c.dev.handle, uint8(num))
			return nil, fmt.Errorf("failed to set alternate config %d on interface %d of %s: %v", alt, num, c, err)
		}
	}

	c.claimed[num] = true
	return &Interface{
		Setting: *altInfo,
		config:  c,
	}, nil
}
