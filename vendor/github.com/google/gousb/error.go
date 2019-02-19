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
)

// #include <libusb.h>
import "C"

// Error is an error code from a USB operation. See the list of Error constants below.
type Error C.int

// Error implements the error interface.
func (e Error) Error() string {
	return fmt.Sprintf("libusb: %s [code %d]", errorString[e], e)
}

func fromErrNo(errno C.int) error {
	err := Error(errno)
	if err == Success {
		return nil
	}
	return err
}

// Defined result codes.
const (
	Success           Error = C.LIBUSB_SUCCESS
	ErrorIO           Error = C.LIBUSB_ERROR_IO
	ErrorInvalidParam Error = C.LIBUSB_ERROR_INVALID_PARAM
	ErrorAccess       Error = C.LIBUSB_ERROR_ACCESS
	ErrorNoDevice     Error = C.LIBUSB_ERROR_NO_DEVICE
	ErrorNotFound     Error = C.LIBUSB_ERROR_NOT_FOUND
	ErrorBusy         Error = C.LIBUSB_ERROR_BUSY
	ErrorTimeout      Error = C.LIBUSB_ERROR_TIMEOUT
	// ErrorOverflow indicates that the device tried to send more data than was
	// requested and that could fit in the packet buffer.
	ErrorOverflow     Error = C.LIBUSB_ERROR_OVERFLOW
	ErrorPipe         Error = C.LIBUSB_ERROR_PIPE
	ErrorInterrupted  Error = C.LIBUSB_ERROR_INTERRUPTED
	ErrorNoMem        Error = C.LIBUSB_ERROR_NO_MEM
	ErrorNotSupported Error = C.LIBUSB_ERROR_NOT_SUPPORTED
	ErrorOther        Error = C.LIBUSB_ERROR_OTHER
)

var errorString = map[Error]string{
	Success:           "success",
	ErrorIO:           "i/o error",
	ErrorInvalidParam: "invalid param",
	ErrorAccess:       "bad access",
	ErrorNoDevice:     "no device",
	ErrorNotFound:     "not found",
	ErrorBusy:         "device or resource busy",
	ErrorTimeout:      "timeout",
	ErrorOverflow:     "overflow",
	ErrorPipe:         "pipe error",
	ErrorInterrupted:  "interrupted",
	ErrorNoMem:        "out of memory",
	ErrorNotSupported: "not supported",
	ErrorOther:        "unknown error",
}

// TransferStatus contains information about the result of a transfer.
type TransferStatus uint8

// Defined Transfer status values.
const (
	TransferCompleted TransferStatus = C.LIBUSB_TRANSFER_COMPLETED
	TransferError     TransferStatus = C.LIBUSB_TRANSFER_ERROR
	TransferTimedOut  TransferStatus = C.LIBUSB_TRANSFER_TIMED_OUT
	TransferCancelled TransferStatus = C.LIBUSB_TRANSFER_CANCELLED
	TransferStall     TransferStatus = C.LIBUSB_TRANSFER_STALL
	TransferNoDevice  TransferStatus = C.LIBUSB_TRANSFER_NO_DEVICE
	TransferOverflow  TransferStatus = C.LIBUSB_TRANSFER_OVERFLOW
)

var transferStatusDescription = map[TransferStatus]string{
	TransferCompleted: "transfer completed without error",
	TransferError:     "transfer failed",
	TransferTimedOut:  "transfer timed out",
	TransferCancelled: "transfer was cancelled",
	TransferStall:     "halt condition detected (endpoint stalled) or control request not supported",
	TransferNoDevice:  "device was disconnected",
	TransferOverflow:  "device sent more data than requested",
}

// String returns a human-readable transfer status.
func (ts TransferStatus) String() string {
	return transferStatusDescription[ts]
}

// Error implements the error interface.
func (ts TransferStatus) Error() string {
	return ts.String()
}
