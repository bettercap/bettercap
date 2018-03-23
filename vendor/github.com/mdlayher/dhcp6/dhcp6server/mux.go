package dhcp6server

import (
	"sync"

	"github.com/mdlayher/dhcp6"
)

// ServeMux is a DHCP request multiplexer, which implements Handler.  ServeMux
// matches handlers based on their MessageType, enabling different handlers
// to be used for different types of DHCP messages.  ServeMux can be helpful
// for structuring your application, but may not be needed for very simple
// DHCP servers.
type ServeMux struct {
	mu sync.RWMutex
	m  map[dhcp6.MessageType]Handler
}

// NewServeMux creates a new ServeMux which is ready to accept Handlers.
func NewServeMux() *ServeMux {
	return &ServeMux{
		m: make(map[dhcp6.MessageType]Handler),
	}
}

// ServeDHCP implements Handler for ServeMux, and serves a DHCP request using
// the appropriate handler for an input Request's MessageType.  If the
// MessageType does not match a valid Handler, ServeDHCP does not invoke any
// handlers, ignoring a client's request.
func (mux *ServeMux) ServeDHCP(w ResponseSender, r *Request) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()
	h, ok := mux.m[r.MessageType]
	if !ok {
		return
	}

	h.ServeDHCP(w, r)
}

// Handle registers a MessageType and Handler with a ServeMux, so that
// future requests with that MessageType will invoke the Handler.
func (mux *ServeMux) Handle(mt dhcp6.MessageType, handler Handler) {
	mux.mu.Lock()
	mux.m[mt] = handler
	mux.mu.Unlock()
}

// HandleFunc registers a MessageType and function as a HandlerFunc with a
// ServeMux, so that future requests with that MessageType will invoke the
// HandlerFunc.
func (mux *ServeMux) HandleFunc(mt dhcp6.MessageType, handler func(ResponseSender, *Request)) {
	mux.Handle(mt, HandlerFunc(handler))
}
