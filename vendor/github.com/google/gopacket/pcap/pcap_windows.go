// Copyright 2012 Google, Inc. All rights reserved.
// Copyright 2009-2011 Andreas Krennmair. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

package pcap

/*
#include <pcap.h>
*/
import "C"

import (
	"errors"
	"os"
	"runtime"
	"unsafe"
)

func (p *Handle) setNonBlocking() error {
	// do nothing
	return nil
}

// waitForPacket waits for a packet or for the timeout to expire.
func (p *Handle) waitForPacket() {
	// can't use select() so instead just switch goroutines
	runtime.Gosched()
}

// openOfflineFile returns contents of input file as a *Handle.
func openOfflineFile(file *os.File) (handle *Handle, err error) {
	buf := (*C.char)(C.calloc(errorBufferSize, 1))
	defer C.free(unsafe.Pointer(buf))
	cf := C.intptr_t(file.Fd())

	cptr := C.pcap_hopen_offline(cf, buf)
	if cptr == nil {
		return nil, errors.New(C.GoString(buf))
	}
	return &Handle{cptr: cptr}, nil
}
