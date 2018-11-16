// Package casm implements raw CASM hosts
package casm

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/lthibault/casm/pkg/net"
)

func init() { rand.Seed(time.Now().UTC().UnixNano()) }

// PeerID uniquely identifies a host instance
type PeerID = net.PeerID

// NewID produces a random PeerID
func NewID() PeerID              { return PeerID(rand.Uint64()) }
func (id PeerID) String() string { return fmt.Sprintf("%016x", uint64(id)) }

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
