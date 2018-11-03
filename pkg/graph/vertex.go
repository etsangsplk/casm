// Package graph implements the CASM expander graph model
package graph

import (
	"context"

	net "github.com/libp2p/go-libp2p-net"
	casm "github.com/lthibault/casm/pkg"
	ma "github.com/multiformats/go-multiaddr"
)

// compile-time type constraint
var _ Vertex = &vertex{}

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

// vertex is a concrete Vertex
type vertex struct {
	h    casm.Host
	b    *broadcast
	k, l uint8
}

// New Vertex
func New(h casm.Host, opt ...Option) (v Vertex, err error) {
	vtx := &vertex{h: h, b: newBroadcaster(h.Addr())}

	vtx.h.Stream().Register(pathEdge, casm.HandlerFunc(vtx.handleEdge))

	for _, o := range append([]Option{OptDefault()}, opt...) {
		if err = o(vtx); err != nil {
			break
		}
	}

	v = vtx
	return
}

// Addr returns the Vertex's network address
func (v vertex) Addr() casm.Addr { return v.h.Addr() }

// Context to which the Vertex's underlying host is bound
func (v vertex) Context() context.Context { return v.h.Context() }

// Message provides an interface to broadcast/pubsub functionality
func (v vertex) Message() Broadcaster { return v.b }

// Edge provides an interface for connecting to peeers
func (v *vertex) Edge() Neighborhood { return v }

// In returns true if the vertex has an edge to the specified peer
func (v vertex) In(id casm.IDer) (ok bool) {
	_, ok = v.h.PeerAddr(id)
	return
}

// Lease an edge slot to the specified peer
func (v vertex) Lease(c context.Context, a casm.Addresser) error {
	if v.In(a.Addr()) {
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
func (v vertex) Evict(id casm.IDer) {
	panic("Evict NOT IMPLEMENTED")
}

func (v vertex) handleEdge(s casm.Stream) {
	panic("handleEdge NOT IMPLEMENTED")
}

/* Implement NetHook */

// Listen is called when the host begins listening
func (v vertex) Listen(net.Network, ma.Multiaddr) {}

// ListenClose is called when the host stops listening
func (v vertex) ListenClose(net.Network, ma.Multiaddr) {}

// Connected is called when a connection is opened
func (v vertex) Connected(net.Network, net.Conn) {}

// Disconnected is called when a connection is closed
func (v vertex) Disconnected(net.Network, net.Conn) {}

// OpenedStream is called when a stream is opened
func (v vertex) OpenedStream(net.Network, net.Stream) {}

// ClosedStream is called when a stream is closed
func (v vertex) ClosedStream(net.Network, net.Stream) {}
