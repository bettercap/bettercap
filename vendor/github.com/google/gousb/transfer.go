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
	"errors"
	"runtime"
	"sync"
)

type usbTransfer struct {
	// mu protects the transfer state.
	mu sync.Mutex
	// xfer is the allocated libusb_transfer.
	xfer *libusbTransfer
	// buf is the buffer allocated for the transfer. The underlying memory
	// is allocated by the C code, both buf and xfer.buffer point to the same
	// memory.
	buf []byte
	// done is blocking until the transfer is complete and data and transfer
	// status are available.
	done chan struct{}
	// submitted is true if submit() was called on this transfer.
	submitted bool
	// ctx is the Context that created this transfer.
	ctx *Context
}

// submits the transfer. After submit() the transfer is in flight and is owned by libusb.
// It's not safe to access the contents of the transfer until wait() returns.
// Once wait() returns, it's ok to re-use the same transfer structure by calling submit() again.
func (t *usbTransfer) submit() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.submitted {
		return errors.New("transfer was already submitted and is not finished yet")
	}
	if err := t.ctx.libusb.submit(t.xfer); err != nil {
		return err
	}
	t.submitted = true
	return nil
}

// waits for libusb to signal the release of transfer data.
// After wait returns, the transfer contents are safe to access
// via t.buf. The number returned by wait indicates how many bytes
// of the buffer were read or written by libusb, and it can be
// smaller than the length of t.buf.
func (t *usbTransfer) wait(ctx context.Context) (n int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.submitted {
		return 0, nil
	}
	select {
	case <-ctx.Done():
		t.ctx.libusb.cancel(t.xfer)
		// after the transfer is cancelled, it will run a callback
		// that triggers the activation of t.done.
		<-t.done
	case <-t.done:
	}
	t.submitted = false
	n, status := t.ctx.libusb.data(t.xfer)
	if status != TransferCompleted {
		return n, status
	}
	return n, err
}

// cancel aborts a submitted transfer. The transfer is cancelled
// asynchronously and the user still needs to wait() to return.
func (t *usbTransfer) cancel() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.submitted {
		return nil
	}
	err := t.ctx.libusb.cancel(t.xfer)
	if err == ErrorNotFound {
		// transfer already completed
		return nil
	}
	return err
}

// free releases the memory allocated for the transfer.
// free should be called only if the transfer is not used by libusb,
// i.e. it should not be called after submit() and before wait() returns.
func (t *usbTransfer) free() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.submitted {
		return errors.New("free() cannot be called on a submitted transfer until wait() returns")
	}
	if t.xfer == nil {
		return nil
	}
	t.ctx.libusb.free(t.xfer)
	t.xfer = nil
	t.buf = nil
	t.done = nil
	return nil
}

// data returns the slice containing transfer buffer.
func (t *usbTransfer) data() []byte {
	return t.buf
}

// newUSBTransfer allocates a new transfer structure and a new buffer for
// communication with a given device/endpoint.
func newUSBTransfer(ctx *Context, dev *libusbDevHandle, ei *EndpointDesc, bufLen int) (*usbTransfer, error) {
	var isoPackets, isoPktSize int
	if ei.TransferType == TransferTypeIsochronous {
		isoPktSize = ei.MaxPacketSize
		if bufLen < isoPktSize {
			isoPktSize = bufLen
		}
		if isoPktSize > 0 {
			isoPackets = bufLen / isoPktSize
		} else {
			isoPackets = 1
		}
		debug.Printf("New isochronous transfer - buffer length %d, using %d packets of %d bytes each", bufLen, isoPackets, isoPktSize)
	}

	done := make(chan struct{}, 1)
	xfer, err := ctx.libusb.alloc(dev, ei, isoPackets, bufLen, done)
	if err != nil {
		return nil, err
	}

	if ei.TransferType == TransferTypeIsochronous {
		ctx.libusb.setIsoPacketLengths(xfer, uint32(isoPktSize))
	}

	t := &usbTransfer{
		xfer: xfer,
		buf:  ctx.libusb.buffer(xfer),
		done: done,
		ctx:  ctx,
	}
	runtime.SetFinalizer(t, func(t *usbTransfer) {
		t.cancel()
		t.wait(context.Background())
		t.free()
	})
	return t, nil
}
