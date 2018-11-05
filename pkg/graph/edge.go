package graph

import (
	"context"
	"io"
	"sync"
	"time"

	casm "github.com/lthibault/casm/pkg"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Pipe is a bidirectional stream of bytes
type Pipe interface {
	io.ReadWriteCloser
	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// Edge is a bidirectional network connection between two vertices
type Edge interface {
	Context() context.Context
	RemotePeer() casm.PeerID
	Pipe() Pipe
}

type streamOpener interface {
	Open(context.Context, casm.Addresser, string) (casm.Stream, error)
}

func negotiateEdge(c context.Context, o streamOpener, a casm.Addresser) (Edge, error) {
	var g *errgroup.Group
	var data, ctrl casm.Stream

	g, c = errgroup.WithContext(c)
	g.Go(func() (err error) {
		if data, err = o.Open(c, a, pathEdgeData); err != nil {
			err = errors.Wrap(err, "data stream")
		}
		return
	})
	g.Go(func() (err error) {
		if ctrl, err = o.Open(c, a, pathEdgeCtrl); err != nil {
			err = errors.Wrap(err, "ctrl stream")
		}
		return
	})

	if err := errors.Wrap(g.Wait(), "open stream"); err != nil {
		// If one stream failed to open, close the other
		for _, s := range []casm.Stream{data, ctrl} {
			if s != nil {
				s.Close()
			}
		}
		return nil, err
	}

	return newEdge(data, ctrl), nil
}

func newEdge(data, ctrl casm.Stream) Edge {
	panic("newEdge NOT IMPLEMENTED")
}

type edgeNegotiator struct {
	sync.Mutex
	m map[casm.PeerID]chan casm.Stream
}

func newEdgeNegotiator() *edgeNegotiator {
	return &edgeNegotiator{m: make(map[casm.PeerID]chan casm.Stream)}
}

func (n *edgeNegotiator) Clear(id casm.IDer) {
	n.Lock()
	if ch, ok := n.m[id.ID()]; ok {
		close(ch)
		delete(n.m, id.ID())
	}
	n.Unlock()
}

func (n *edgeNegotiator) maybeInitUnsafe(id casm.PeerID) (ch chan casm.Stream) {
	var ok bool
	if ch, ok = n.m[id]; !ok {
		ch = make(chan casm.Stream)
		n.m[id] = ch
	}
	return
}

func (n *edgeNegotiator) ProvideDataStream(id casm.IDer) chan<- casm.Stream {
	n.Lock()
	defer n.Unlock()
	return n.maybeInitUnsafe(id.ID())
}

func (n *edgeNegotiator) WaitDataStream(id casm.IDer) <-chan casm.Stream {
	n.Lock()
	defer n.Unlock()
	return n.maybeInitUnsafe(id.ID())
}
