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
	"sort"
)

// InterfaceDesc contains information about a USB interface, extracted from
// the descriptor.
type InterfaceDesc struct {
	// Number is the number of this interface, a zero-based index in the array
	// of interfaces supported by the device configuration.
	Number int
	// AltSettings is a list of alternate settings supported by the interface.
	AltSettings []InterfaceSetting
}

// String returns a human-readable description of the interface descriptor and
// its alternate settings.
func (i InterfaceDesc) String() string {
	return fmt.Sprintf("Interface %d (%d alternate settings)", i.Number, len(i.AltSettings))
}

// InterfaceSetting contains information about a USB interface with a particular
// alternate setting, extracted from the descriptor.
type InterfaceSetting struct {
	// Number is the number of this interface, the same as in InterfaceDesc.
	Number int
	// Alternate is the number of this alternate setting.
	Alternate int
	// Class is the USB-IF (Implementers Forum) class code, as defined by the USB spec.
	Class Class
	// SubClass is the USB-IF (Implementers Forum) subclass code, as defined by the USB spec.
	SubClass Class
	// Protocol is USB protocol code, as defined by the USB spe.c
	Protocol Protocol
	// Endpoints enumerates the endpoints available on this interface with
	// this alternate setting.
	Endpoints map[EndpointAddress]EndpointDesc

	iInterface int // index of a string descriptor describing this interface.
}

func (a InterfaceSetting) sortedEndpointIds() []string {
	var eps []string
	for _, ei := range a.Endpoints {
		eps = append(eps, fmt.Sprintf("%s(%d,%s)", ei.Address, ei.Number, ei.Direction))
	}
	sort.Strings(eps)
	return eps
}

// String returns a human-readable description of the particular
// alternate setting of an interface.
func (a InterfaceSetting) String() string {
	return fmt.Sprintf("Interface %d alternate setting %d (available endpoints: %v)", a.Number, a.Alternate, a.sortedEndpointIds())
}

// Interface is a representation of a claimed interface with a particular setting.
// To access device endpoints use InEndpoint() and OutEndpoint() methods.
// The interface should be Close()d after use.
type Interface struct {
	Setting InterfaceSetting

	config *Config
}

func (i *Interface) String() string {
	return fmt.Sprintf("%s,if=%d,alt=%d", i.config, i.Setting.Number, i.Setting.Alternate)
}

// Close releases the interface.
func (i *Interface) Close() {
	if i.config == nil {
		return
	}
	i.config.dev.ctx.libusb.release(i.config.dev.handle, uint8(i.Setting.Number))
	i.config.mu.Lock()
	defer i.config.mu.Unlock()
	delete(i.config.claimed, i.Setting.Number)
	i.config = nil
}

func (i *Interface) openEndpoint(epAddr EndpointAddress) (*endpoint, error) {
	var ep EndpointDesc
	ep, ok := i.Setting.Endpoints[epAddr]
	if !ok {
		return nil, fmt.Errorf("%s does not have endpoint with address %s. Available endpoints: %v", i, epAddr, i.Setting.sortedEndpointIds())
	}
	return &endpoint{
		InterfaceSetting: i.Setting,
		Desc:             ep,
		h:                i.config.dev.handle,
		ctx:              i.config.dev.ctx,
	}, nil
}

// InEndpoint prepares an IN endpoint for transfer.
func (i *Interface) InEndpoint(epNum int) (*InEndpoint, error) {
	if i.config == nil {
		return nil, fmt.Errorf("InEndpoint(%d) called on %s after Close", epNum, i)
	}
	ep, err := i.openEndpoint(EndpointAddress(0x80 | epNum))
	if err != nil {
		return nil, err
	}
	return &InEndpoint{
		endpoint: ep,
	}, nil
}

// OutEndpoint prepares an OUT endpoint for transfer.
func (i *Interface) OutEndpoint(epNum int) (*OutEndpoint, error) {
	if i.config == nil {
		return nil, fmt.Errorf("OutEndpoint(%d) called on %s after Close", epNum, i)
	}
	ep, err := i.openEndpoint(EndpointAddress(epNum))
	if err != nil {
		return nil, err
	}
	return &OutEndpoint{
		endpoint: ep,
	}, nil
}
