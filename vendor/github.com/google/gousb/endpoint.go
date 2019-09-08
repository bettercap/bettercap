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
	"context"
	"fmt"
	"strings"
	"time"
)

// EndpointAddress is a unique identifier for the endpoint, combining the endpoint number and direction.
type EndpointAddress uint8

// String implements the Stringer interface.
func (a EndpointAddress) String() string {
	return fmt.Sprintf("0x%02x", uint8(a))
}

// EndpointDesc contains the information about an interface endpoint, extracted
// from the descriptor.
type EndpointDesc struct {
	// Address is the unique identifier of the endpoint within the interface.
	Address EndpointAddress
	// Number represents the endpoint number. Note that the endpoint number is different from the
	// address field in the descriptor - address 0x82 means endpoint number 2,
	// with endpoint direction IN.
	// The device can have up to two endpoints with the same number but with
	// different directions.
	Number int
	// Direction defines whether the data is flowing IN or OUT from the host perspective.
	Direction EndpointDirection
	// MaxPacketSize is the maximum USB packet size for a single frame/microframe.
	MaxPacketSize int
	// TransferType defines the endpoint type - bulk, interrupt, isochronous.
	TransferType TransferType
	// PollInterval is the maximum time between transfers for interrupt and isochronous transfer,
	// or the NAK interval for a control transfer. See endpoint descriptor bInterval documentation
	// in the USB spec for details.
	PollInterval time.Duration
	// IsoSyncType is the isochronous endpoint synchronization type, as defined by USB spec.
	IsoSyncType IsoSyncType
	// UsageType is the isochronous or interrupt endpoint usage type, as defined by USB spec.
	UsageType UsageType
}

// String returns the human-readable description of the endpoint.
func (e EndpointDesc) String() string {
	ret := make([]string, 0, 3)
	ret = append(ret, fmt.Sprintf("ep #%d %s (address %s) %s", e.Number, e.Direction, e.Address, e.TransferType))
	switch e.TransferType {
	case TransferTypeIsochronous:
		ret = append(ret, fmt.Sprintf("- %s %s", e.IsoSyncType, e.UsageType))
	case TransferTypeInterrupt:
		ret = append(ret, fmt.Sprintf("- %s", e.UsageType))
	}
	ret = append(ret, fmt.Sprintf("[%d bytes]", e.MaxPacketSize))
	return strings.Join(ret, " ")
}

type endpoint struct {
	h *libusbDevHandle

	InterfaceSetting
	Desc EndpointDesc

	ctx *Context
}

// String returns a human-readable description of the endpoint.
func (e *endpoint) String() string {
	return e.Desc.String()
}

func (e *endpoint) transfer(ctx context.Context, buf []byte) (int, error) {
	t, err := newUSBTransfer(e.ctx, e.h, &e.Desc, len(buf))
	if err != nil {
		return 0, err
	}
	defer t.free()
	if e.Desc.Direction == EndpointDirectionOut {
		copy(t.data(), buf)
	}

	if err := t.submit(); err != nil {
		return 0, err
	}

	n, err := t.wait(ctx)
	if e.Desc.Direction == EndpointDirectionIn {
		copy(buf, t.data())
	}
	if err != nil {
		return n, err
	}
	return n, nil
}

// InEndpoint represents an IN endpoint open for transfer.
// InEndpoint implements the io.Reader interface.
// For high-throughput transfers, consider creating a bufffered read stream
// through InEndpoint.ReadStream.
type InEndpoint struct {
	*endpoint
}

// Read reads data from an IN endpoint. Read returns number of bytes obtained
// from the endpoint. Read may return non-zero length even if
// the returned error is not nil (partial read).
func (e *InEndpoint) Read(buf []byte) (int, error) {
	return e.transfer(context.Background(), buf)
}

// ReadContext reads data from an IN endpoint. ReadContext returns number of
// bytes obtained from the endpoint. ReadContext may return non-zero length
// even if the returned error is not nil (partial read).
// The passed context can be used to control the cancellation of the read. If
// the context is cancelled, ReadContext will cancel the underlying transfers,
// resulting in TransferCancelled error.
func (e *InEndpoint) ReadContext(ctx context.Context, buf []byte) (int, error) {
	return e.transfer(ctx, buf)
}

// OutEndpoint represents an OUT endpoint open for transfer.
type OutEndpoint struct {
	*endpoint
}

// Write writes data to an OUT endpoint. Write returns number of bytes comitted
// to the endpoint. Write may return non-zero length even if the returned error
// is not nil (partial write).
func (e *OutEndpoint) Write(buf []byte) (int, error) {
	return e.transfer(context.Background(), buf)
}

// WriteContext writes data to an OUT endpoint. WriteContext returns number of
// bytes comitted to the endpoint. WriteContext may return non-zero length even
// if the returned error is not nil (partial write).
// The passed context can be used to control the cancellation of the write. If
// the context is cancelled, WriteContext will cancel the underlying transfers,
// resulting in TransferCancelled error.
func (e *OutEndpoint) WriteContext(ctx context.Context, buf []byte) (int, error) {
	return e.transfer(ctx, buf)
}
