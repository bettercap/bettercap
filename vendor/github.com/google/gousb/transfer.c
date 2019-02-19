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

#include <libusb.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void print_xfer(struct libusb_transfer *xfer);
void xferCallback(struct libusb_transfer*);

int submit(struct libusb_transfer *xfer) {
	xfer->callback = (libusb_transfer_cb_fn)(&xferCallback);
	xfer->status = -1;
	return libusb_submit_transfer(xfer);
}

void print_xfer(struct libusb_transfer *xfer) {
	int i;

	printf("Transfer:\n");
	printf("  dev_handle:   %p\n", xfer->dev_handle);
	printf("  flags:        %08x\n", xfer->flags);
	printf("  endpoint:     %x\n", xfer->endpoint);
	printf("  type:         %x\n", xfer->type);
	printf("  timeout:      %dms\n", xfer->timeout);
	printf("  status:       %x\n", xfer->status);
	printf("  length:       %d (act: %d)\n", xfer->length, xfer->actual_length);
	printf("  callback:     %p\n", xfer->callback);
	printf("  user_data:    %p\n", xfer->user_data);
	printf("  buffer:       %p\n", xfer->buffer);
	printf("  num_iso_pkts: %d\n", xfer->num_iso_packets);
	printf("  packets:\n");
	for (i = 0; i < xfer->num_iso_packets; i++) {
		printf("    [%04d] %d (act: %d) %x\n", i,
			xfer->iso_packet_desc[i].length,
			xfer->iso_packet_desc[i].actual_length,
			xfer->iso_packet_desc[i].status);
	}
}

// compact the data in an isochronous transfer. The contents of individual
// iso packets are shifted left, so that no gaps are left between them.
// Status is set to the first non-zero status of an iso packet.
int gousb_compact_iso_data(struct libusb_transfer *xfer, unsigned char *status) {
	int i;
	int sum = 0;
	unsigned char *in = xfer->buffer;
	unsigned char *out = xfer->buffer;
	for (i = 0; i < xfer->num_iso_packets; i++) {
		struct libusb_iso_packet_descriptor pkt = xfer->iso_packet_desc[i];
		if (pkt.status != 0) {
		    *status = pkt.status;
			break;
		}
		// Copy the data
		int len = pkt.actual_length;
		memmove(out, in, len);
		// Increment offsets
		sum += len;
		in += pkt.length;
		out += len;
	}
	return sum;
}

// allocates a libusb transfer and a buffer for packet data.
struct libusb_transfer *gousb_alloc_transfer_and_buffer(int bufLen, int isoPackets) {
        struct libusb_transfer *xfer = libusb_alloc_transfer(isoPackets);
        if (xfer == NULL) {
                return NULL;
        }
        xfer->buffer = (unsigned char*)malloc(bufLen);
        if (xfer->buffer == NULL) {
                libusb_free_transfer(xfer);
                return NULL;
        }
        xfer->length = bufLen;
        return xfer;
}

// frees a libusb transfer and its buffer. The buffer of the given
// libusb_transfer must have been allocated with alloc_transfer_and_buffer.
void gousb_free_transfer_and_buffer(struct libusb_transfer *xfer) {
        free(xfer->buffer);
        xfer->length = 0;
        libusb_free_transfer(xfer);
}
