// Package graph implements the CASM expander graph model
package graph

import (
	"context"

	casm "github.com/lthibault/casm/pkg"
)

// compile-time type constraint
var _ Vertex = &V{}

// Vertex in the expander graph
type Vertex interface {
	Addr() casm.Addr
	Context() context.Context
	Message() Broadcaster
	Edge() Neighborhood
}

// V is a concrete Vertex
type V struct {
	h    casm.Host
	b    *broadcast
	k, l uint8
}

// New V
func New(h casm.Host, opt ...Option) (v *V, err error) {
	v = &V{h: h, b: newBroadcaster(h.Addr())}
	for _, o := range append([]Option{OptDefault()}, opt...) {
		if _, err = o(v); err != nil {
			break
		}
	}

	return
}

// Addr returns the Vertex's network address
func (v V) Addr() casm.Addr { return v.h.Addr() }

// Context to which the Vertex's underlying host is bound
func (v V) Context() context.Context { return v.h.Context() }

// Message provides an interface to broadcast/pubsub functionality
func (v V) Message() Broadcaster { return v.b }

// Edge provides an interface for connecting to peeers
func (v *V) Edge() Neighborhood { return v }

// Lease an edge slot to the specified peer
func (v V) Lease(a casm.Addr) {
	panic("Lease NOT IMPLEMENTED")
}

// Evict the specified peer from the vertex, closing all connections
func (v V) Evict(id casm.IDer) {
	panic("Evict NOT IMPLEMENTED")
}
