// Package casm implements raw CASM hosts
package casm

import (
	"context"
)

// Network manages raw connections
type Network interface {
	Connect(context.Context, Addresser) error
	Disconnect(IDer)
}

// StreamManager manages streams, which are multiplexed on top of raw connections
type StreamManager interface {
	Register(string, Handler)
	Unregister(string)
	Open(context.Context, Addresser, string) (Stream, error)
}

// Host is a logical machine in a compute cluster.  It acts both as a server and
// a client.  In the CASM expander-graph model, it is a vertex.
type Host interface {
	Context() context.Context
	Addr() Addr
	Network() Network
	Stream() StreamManager
}
