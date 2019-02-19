// Copyright 2017 the gousb Authors.  All rights reserved.
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
	"io"
)

type transferIntf interface {
	submit() error
	cancel() error
	wait(context.Context) (int, error)
	free() error
	data() []byte
}

type stream struct {
	// a fifo of USB transfers.
	transfers chan transferIntf
	// err is the first encountered error, returned to the user.
	err error
	// finished is true if transfers has been already closed.
	finished bool
}

func (s *stream) gotError(err error) {
	if s.err == nil {
		s.err = err
	}
}

func (s *stream) noMore() {
	if !s.finished {
		close(s.transfers)
		s.finished = true
	}
}

func (s *stream) submitAll() {
	count := len(s.transfers)
	var all []transferIntf
	for i := 0; i < count; i++ {
		all = append(all, <-s.transfers)
	}
	for _, t := range all {
		if err := t.submit(); err != nil {
			t.free()
			s.gotError(err)
			s.noMore()
			return
		}
		s.transfers <- t
	}
	return
}

func (s *stream) flushRemaining() {
	s.noMore()
	for t := range s.transfers {
		t.cancel()
		t.wait(context.Background())
		t.free()
	}
}

func (s *stream) done() {
	if s.err == nil {
		close(s.transfers)
	}
}

// ReadStream is a buffer that tries to prefetch data from the IN endpoint,
// reducing the latency between subsequent Read()s.
// ReadStream keeps prefetching data until Close() is called or until
// an error is encountered. After Close(), the buffer might still have
// data left from transfers that were initiated before Close. Read()ing
// from the ReadStream will keep returning available data. When no more
// data is left, io.EOF is returned.
type ReadStream struct {
	s *stream
	// current holds the last transfer to return.
	current transferIntf
	// total/used are the number of all/used bytes in the current transfer.
	total, used int
}

// Read reads data from the transfer stream.
// The data will come from at most a single transfer, so the returned number
// might be smaller than the length of p.
// After a non-nil error is returned, all subsequent attempts to read will
// return io.ErrClosedPipe.
// Read cannot be called concurrently with other Read, ReadContext
// or Close.
func (r *ReadStream) Read(p []byte) (int, error) {
	return r.ReadContext(context.Background(), p)
}

// ReadContext reads data from the transfer stream.
// The data will come from at most a single transfer, so the returned number
// might be smaller than the length of p.
// After a non-nil error is returned, all subsequent attempts to read will
// return io.ErrClosedPipe.
// ReadContext cannot be called concurrently with other Read, ReadContext
// or Close.
// The context passed controls the cancellation of this particular read
// operation within the stream. The semantics is identical to
// Endpoint.ReadContext.
func (r *ReadStream) ReadContext(ctx context.Context, p []byte) (int, error) {
	if r.s.transfers == nil {
		return 0, io.ErrClosedPipe
	}
	if r.current == nil {
		t, ok := <-r.s.transfers
		if !ok {
			// no more transfers in flight
			r.s.transfers = nil
			return 0, r.s.err
		}
		n, err := t.wait(ctx)
		if err != nil {
			// wait error aborts immediately, all remaining data is invalid.
			t.free()
			r.s.flushRemaining()
			r.s.transfers = nil
			return n, err
		}
		r.current = t
		r.total = n
		r.used = 0
	}
	use := r.total - r.used
	if use > len(p) {
		use = len(p)
	}
	copy(p, r.current.data()[r.used:r.used+use])
	r.used += use
	if r.used == r.total {
		if r.s.err == nil {
			if err := r.current.submit(); err == nil {
				// guaranteed to not block, len(transfers) == number of allocated transfers
				r.s.transfers <- r.current
			} else {
				r.s.gotError(err)
				r.s.noMore()
			}
		}
		if r.s.err != nil {
			r.current.free()
		}
		r.current = nil
	}
	return use, nil
}

// Close signals that the transfer should stop. After Close is called,
// subsequent Read()s will return data from all transfers that were already
// in progress before returning an io.EOF error, unless another error
// was encountered earlier.
// Close cannot be called concurrently with Read.
func (r *ReadStream) Close() error {
	if r.s.transfers == nil {
		return nil
	}
	r.s.gotError(io.EOF)
	r.s.noMore()
	return nil
}

// WriteStream is a buffer that will send data asynchronously, reducing
// the latency between subsequent Write()s.
type WriteStream struct {
	s     *stream
	total int
}

// Write sends the data to the endpoint. Write returning a nil error doesn't
// mean that data was written to the device, only that it was written to the
// buffer. Only a call to Close() that returns nil error guarantees that
// all transfers have succeeded.
// If the slice passed to Write does not align exactly with the transfer
// buffer size (as declared in a call to NewStream), the last USB transfer
// of this Write will be sent with less data than the full buffer.
// After a non-nil error is returned, all subsequent attempts to write will
// return io.ErrClosedPipe.
// If Write encounters an error when preparing the transfer, the stream
// will still try to complete any pending transfers. The total number
// of bytes successfully written can be retrieved through a Written()
// call after Close() has returned.
// Write cannot be called concurrently with another Write, Written or Close.
func (w *WriteStream) Write(p []byte) (int, error) {
	return w.WriteContext(context.Background(), p)
}

// WriteContext sends the data to the endpoint. Write returning a nil error doesn't
// mean that data was written to the device, only that it was written to the
// buffer. Only a call to Close() that returns nil error guarantees that
// all transfers have succeeded.
// If the slice passed to WriteContext does not align exactly with the transfer
// buffer size (as declared in a call to NewStream), the last USB transfer
// of this Write will be sent with less data than the full buffer.
// After a non-nil error is returned, all subsequent attempts to write will
// return io.ErrClosedPipe.
// If WriteContext encounters an error when preparing the transfer, the stream
// will still try to complete any pending transfers. The total number
// of bytes successfully written can be retrieved through a Written()
// call after Close() has returned.
// WriteContext cannot be called concurrently with another Write, WriteContext,
// Written, Close or CloseContext.
func (w *WriteStream) WriteContext(ctx context.Context, p []byte) (int, error) {
	if w.s.transfers == nil || w.s.err != nil {
		return 0, io.ErrClosedPipe
	}
	written := 0
	all := len(p)
	for written < all {
		t := <-w.s.transfers
		n, err := t.wait(ctx) // unsubmitted transfers will return 0 bytes and no error
		w.total += n
		if err != nil {
			t.free()
			w.s.gotError(err)
			// This branch is used only after all the transfers were set in flight.
			// That means all transfers left in the queue are in flight.
			// They must be ignored, since this wait() failed.
			w.s.flushRemaining()
			return written, err
		}
		use := all - written
		if max := len(t.data()); use > max {
			use = max
		}
		copy(t.data(), p[written:written+use])
		if err := t.submit(); err != nil {
			t.free()
			w.s.gotError(err)
			// Even though this submit failed, all the transfers in flight are still valid.
			// Don't flush remaining transfers.
			// We won't submit any more transfers.
			w.s.noMore()
			return written, err
		}
		written += use
		w.s.transfers <- t // guaranteed non blocking
	}
	return written, nil
}

// Close signals end of data to write. Close blocks until all transfers
// that were sent are finished. The error returned by Close is the first
// error encountered during writing the entire stream (if any).
// Close returning nil indicates all transfers completed successfully.
// After Close, the total number of bytes successfully written can be
// retrieved using Written().
// Close may not be called concurrently with Write, Close or Written.
func (w *WriteStream) Close() error {
	return w.CloseContext(context.Background())
}

// Close signals end of data to write. Close blocks until all transfers
// that were sent are finished. The error returned by Close is the first
// error encountered during writing the entire stream (if any).
// Close returning nil indicates all transfers completed successfully.
// After Close, the total number of bytes successfully written can be
// retrieved using Written().
// Close may not be called concurrently with Write, Close or Written.
// CloseContext
func (w *WriteStream) CloseContext(ctx context.Context) error {
	if w.s.transfers == nil {
		return io.ErrClosedPipe
	}
	w.s.noMore()
	for t := range w.s.transfers {
		n, err := t.wait(ctx)
		w.total += n
		t.free()
		if err != nil {
			w.s.gotError(err)
			w.s.flushRemaining()
		}
		t.free()
	}
	w.s.transfers = nil
	return w.s.err
}

// Written returns the number of bytes successfully written by the stream.
// Written may be called only after Close() or CloseContext()
// has been called and returned.
func (w *WriteStream) Written() int {
	return w.total
}

func newStream(tt []transferIntf) *stream {
	s := &stream{
		transfers: make(chan transferIntf, len(tt)),
	}
	for _, t := range tt {
		s.transfers <- t
	}
	return s
}
