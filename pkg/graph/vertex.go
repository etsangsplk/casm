// Package graph implements the CASM expander graph model
package graph

import (
	"context"
	"time"

	net "github.com/libp2p/go-libp2p-net"
	casm "github.com/lthibault/casm/pkg"
	ma "github.com/multiformats/go-multiaddr"
)

// compile-time type constraint
var _ Vertex = &vertex{}

const (
	pathEdgeData = "/edge/data"
	pathEdgeCtrl = "/edge/ctrl"
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
	k, l uint8
	b    *broadcast
	en   *edgeNegotiator
}

func newVertex(h casm.Host) *vertex {
	return &vertex{
		h:  h,
		b:  newBroadcaster(h.Addr()),
		en: newEdgeNegotiator(),
	}
}

func (v *vertex) configure(opt []Option) (err error) {
	for _, o := range append([]Option{OptDefault()}, opt...) {
		if err = o(v); err != nil {
			break
		}
	}

	v.h.Stream().Register(pathEdgeData, casm.HandlerFunc(v.initEdgeData))
	v.h.Stream().Register(pathEdgeCtrl, casm.HandlerFunc(v.initEdgeCtrl))
	v.h.Network().Hook().Add(v)

	return
}

// New Vertex
func New(h casm.Host, opt ...Option) (Vertex, error) {
	v := newVertex(h)
	return v, v.configure(opt)
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
	panic("In NOT IMPLEMENTED")
}

// Lease an edge slot to the specified peer
func (v vertex) Lease(c context.Context, a casm.Addresser) error {
	if v.In(a.Addr()) { // TODO:  ensure that Evict() clears the Host's Peerstore
		return nil
	}

	e, err := negotiateEdge(c, v.h.Stream(), a)
	if err == nil {
		v.b.AddEdge(e)
	}

	return err
}

// Evict the specified peer from the vertex, closing all connections
func (v vertex) Evict(id casm.IDer) {
	if e, ok := v.b.RemoveEdge(id); ok {
		e.Close()
	}
}

func (v vertex) initEdgeData(s casm.Stream) {
	c, cancel := context.WithTimeout(s.Context(), time.Second*10)
	defer cancel()
	defer v.en.Clear(s.RemotePeer())

	select {
	case <-s.Context().Done():
	case <-c.Done():
	case v.en.ProvideDataStream(s.RemotePeer()) <- s:
	}
}

func (v vertex) initEdgeCtrl(s casm.Stream) {
	c, cancel := context.WithTimeout(s.Context(), time.Second*10)
	defer cancel()
	defer v.en.Clear(s.RemotePeer())

	select {
	case <-s.Context().Done():
	case <-c.Done():
	case ds, ok := <-v.en.WaitDataStream(s.RemotePeer()):
		if !ok {
			s.Close()
			return
		}

		v.b.AddEdge(newEdge(newStreamGroup(ds, s)))
	}
}

/* Implement NetHook */

// Listen is called when the host begins listening on an addr
func (v vertex) Listen(net.Network, ma.Multiaddr) {
	// incr a counter atomically
	// if the counter was incred FROM 0, start the vertex's state-maintenance logic
}

// ListenClose is called when the host stops listening on an addr
func (v vertex) ListenClose(net.Network, ma.Multiaddr) {
	// decr a counter atomically
	// if the counter is 0, stop the vertex's state-maintenance logic & clean-up
}

// Connected is called when a connection is opened
func (v vertex) Connected(net.Network, net.Conn) {
	// incr a counter that tracks current cardinality
	// trigger notifications of state-change, where necessary
}

// Disconnected is called when a connection is closed
func (v vertex) Disconnected(net.Network, net.Conn) {
	// decr a counter that tracks current cardinality
	// trigger notifications of state-change, where necessary
}

// OpenedStream is called when a stream is opened
func (v vertex) OpenedStream(net.Network, net.Stream) {}

// ClosedStream is called when a stream is closed
func (v vertex) ClosedStream(net.Network, net.Stream) {}
