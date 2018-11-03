// Package graph implements the CASM expander graph model
package graph

import (
	"context"

	casm "github.com/lthibault/casm/pkg"
)

// compile-time type constraint
var _ Vertex = &V{}

const (
	pathEdge = "/edge"
)

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

	v.h.RegisterStreamHandler(pathEdge, casm.HandlerFunc(v.handleEdge))

	for _, o := range append([]Option{OptDefault()}, opt...) {
		if err = o(v); err != nil {
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

// Connected returns true if the vertex has an edge to the specified peer
func (v V) Connected(a casm.Addresser) (ok bool) {
	_, ok = v.h.PeerAddr(a.Addr().Label())
	return
}

// Lease an edge slot to the specified peer
func (v V) Lease(c context.Context, a casm.Addresser) error {
	if v.Connected(a.Addr()) {
		return nil
	}

	// we want to build an edge, then add it to an edgeSet

	// s, err := v.h.OpenStream(c, a, pathEdge)
	// if err != nil {
	// 	return errors.Wrap(err, "open stream")
	// }

	panic("Lease NOT IMPLEMENTED")
}

// Evict the specified peer from the vertex, closing all connections
func (v V) Evict(id casm.IDer) {
	panic("Evict NOT IMPLEMENTED")
}

func (v V) handleEdge(s casm.Stream) {
	panic("handleEdge NOT IMPLEMENTED")
}
