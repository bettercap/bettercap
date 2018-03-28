// +build libnfqueue1

package nfqueue

// This file contains code specific to versions >= 1.0 of libnetfilter_queue

/*
#include <stdio.h>
#include <stdint.h>
#include <arpa/inet.h>
#include <linux/netfilter.h>
#include <libnetfilter_queue/libnetfilter_queue.h>
*/
import "C"

import (
    "log"
    "unsafe"
)

// SetVerdictMark issues a verdict for a packet, but a mark can be set
//
// Every queued packet _must_ have a verdict specified by userspace.
func (p *Payload) SetVerdictMark(verdict int, mark uint32) error {
    log.Printf("Setting verdict for packet %d: %d mark %lx\n",p.Id,verdict,mark)
    C.nfq_set_verdict2(
        p.c_qh,
        C.u_int32_t(p.Id),
        C.u_int32_t(verdict),
        C.u_int32_t(mark),
        0,nil)
    return nil
}

// SetVerdictMarkModified issues a verdict for a packet, but replaces the
// packet with the provided one, and a mark can be set.
//
// Every queued packet _must_ have a verdict specified by userspace.
func (p *Payload) SetVerdictMarkModified(verdict int, mark uint32, data []byte) error {
    log.Printf("Setting verdict for NEW packet %d: %d mark %lx\n",p.Id,verdict,mark)
    C.nfq_set_verdict2(
        p.c_qh,
        C.u_int32_t(p.Id),
        C.u_int32_t(verdict),
        C.u_int32_t(mark),
        C.u_int32_t(len(data)),
        (*C.uchar)(unsafe.Pointer(&data[0])),
    )
    return nil
}
