// Package graph implements the CASM expander graph model
package graph

import (
	"context"
	"time"

	casm "github.com/lthibault/casm/pkg"
	net "github.com/lthibault/casm/pkg/net"
)

// compile-time type constraint
var _ Vertex = &vertex{}

const (
	pathEdgeData = "/edge/data"
	pathEdgeCtrl = "/edge/ctrl"
)

// Vertex in the expander graph
type Vertex interface {
	Addr() net.Addr
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

	v.h.Stream().Register(pathEdgeData, net.HandlerFunc(v.initEdgeData))
	v.h.Stream().Register(pathEdgeCtrl, net.HandlerFunc(v.initEdgeCtrl))

	return
}

// New Vertex
func New(h casm.Host, opt ...Option) (Vertex, error) {
	v := newVertex(h)
	return v, v.configure(opt)
}

// Addr returns the Vertex's network address
func (v vertex) Addr() net.Addr { return v.h.Addr() }

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

func (v vertex) initEdgeData(s net.Stream) {
	c, cancel := context.WithTimeout(s.Context(), time.Second*10)
	defer cancel()
	defer v.en.Clear(s.Endpoint().Remote())

	select {
	case <-s.Context().Done():
	case <-c.Done():
	case v.en.ProvideDataStream(s.Endpoint().Remote()) <- s:
	}
}

func (v vertex) initEdgeCtrl(s net.Stream) {
	c, cancel := context.WithTimeout(s.Context(), time.Second*10)
	defer cancel()
	defer v.en.Clear(s.Endpoint().Remote())

	select {
	case <-s.Context().Done():
	case <-c.Done():
	case ds, ok := <-v.en.WaitDataStream(s.Endpoint().Remote()):
		if !ok {
			s.Close()
			return
		}

		v.b.AddEdge(newEdge(newStreamGroup(ds, s)))
	}
}
