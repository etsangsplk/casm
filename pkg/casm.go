// Package casm implements raw CASM hosts
package casm

import (
	"context"
	"math/rand"
	"time"

	"github.com/lthibault/casm/pkg/net"
)

func init() { rand.Seed(time.Now().UTC().UnixNano()) }

type (
	// PeerID uniquely identifies a host instance
	PeerID = net.PeerID

	// IDer can provide a PeerID
	IDer = net.IDer
)

// NewID produces a random PeerID
func NewID() PeerID { return PeerID(rand.Uint64()) }

// Addresser can provide an Addr
type Addresser interface {
	Addr() net.Addr
}

// Network manages raw connections
type Network interface {
	Connect(context.Context, Addresser) error
	Disconnect(IDer)
}

// StreamManager manages streams, which are multiplexed on top of raw connections
type StreamManager interface {
	Register(string, net.Handler)
	Unregister(string)
	Open(context.Context, Addresser, string) (net.Stream, error)
}

// Host is a logical machine in a compute cluster.  It acts both as a server and
// a client.  In the CASM expander-graph model, it is a vertex.
type Host interface {
	Context() context.Context
	Addr() net.Addr
	Network() Network
	Stream() StreamManager
}
