package graph

import (
	"context"
	"io"
	"sync"
	"time"

	casm "github.com/lthibault/casm/pkg"
	net "github.com/lthibault/casm/pkg/net"
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

// EdgeAPI provides an interface for querying the graph via a connected peer
type EdgeAPI interface {
}

// Edge is a bidirectional network connection between two vertices
type Edge interface {
	io.Closer
	Context() context.Context
	RemotePeer() net.PeerID
	Pipe() Pipe
	API() EdgeAPI
}

type streamOpener interface {
	Open(context.Context, casm.Addresser, string) (net.Stream, error)
}

func negotiateEdge(c context.Context, o streamOpener, a casm.Addresser) (Edge, error) {
	var g *errgroup.Group
	var data, ctrl net.Stream

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
		for _, s := range []net.Stream{data, ctrl} {
			if s != nil {
				s.Close()
			}
		}
		return nil, err
	}

	return newEdge(newStreamGroup(data, ctrl)), nil
}

// type idempotentCloser struct {
// 	sync.Once
// 	err error
// 	io.Closer
// }

// func (ic *idempotentCloser) Close() error {
// 	ic.Do(func() { ic.err = ic.Closer.Close() })
// 	return ic.err
// }

// type multiCloser []io.Closer

// func newMultiCloser(close ...io.Closer) multiCloser {
// 	var mc multiCloser = make([]io.Closer, len(close))
// 	for i, c := range close {
// 		mc[i] = &idempotentCloser{Closer: c}
// 	}
// 	return mc
// }

// func (mc multiCloser) Close() error {
// 	var g errgroup.Group
// 	for _, c := range mc {
// 		g.Go(c.Close)
// 	}
// 	return g.Wait()
// }

// type threadSafeStream struct {
// 	sync.RWMutex
// 	s net.Stream
// }

// type api struct {
// 	s net.Stream
// }

// func newAPI(s net.Stream) *api {
// 	return &api{s: s}
// }

type edge struct {
}

func newEdge(g streamGroup) *edge {
	panic("newEdge NOT IMPLEMENTED")
}

func (e edge) Context() context.Context { panic("function NOT IMPLEMENTED") }
func (e *edge) Pipe() Pipe              { panic("function NOT IMPLEMENTED") }
func (e *edge) API() EdgeAPI            { panic("function NOT IMPLEMENTED") }
func (e edge) RemotePeer() net.PeerID   { panic("function NOT IMPLEMENTED") }

func (e edge) Close() error {
	panic("function NOT IMPLEMENTED")
}

func (e edge) Read(b []byte) (n int, err error)   { panic("function NOT IMPLEMENTED") }
func (e edge) Write(b []byte) (n int, err error)  { panic("function NOT IMPLEMENTED") }
func (e edge) SetDeadline(t time.Time) error      { panic("function NOT IMPLEMENTED") }
func (e edge) SetReadDeadline(t time.Time) error  { panic("function NOT IMPLEMENTED") }
func (e edge) SetWriteDeadline(t time.Time) error { panic("function NOT IMPLEMENTED") }

type edgeNegotiator struct {
	sync.Mutex
	m map[net.PeerID]chan net.Stream
}

func newEdgeNegotiator() *edgeNegotiator {
	return &edgeNegotiator{m: make(map[net.PeerID]chan net.Stream)}
}

func (n *edgeNegotiator) Clear(id casm.IDer) {
	n.Lock()
	if ch, ok := n.m[id.ID()]; ok {
		close(ch)
		delete(n.m, id.ID())
	}
	n.Unlock()
}

func (n *edgeNegotiator) maybeInitUnsafe(id net.PeerID) (ch chan net.Stream) {
	var ok bool
	if ch, ok = n.m[id]; !ok {
		ch = make(chan net.Stream)
		n.m[id] = ch
	}
	return
}

func (n *edgeNegotiator) ProvideDataStream(id casm.IDer) chan<- net.Stream {
	n.Lock()
	defer n.Unlock()
	return n.maybeInitUnsafe(id.ID())
}

func (n *edgeNegotiator) WaitDataStream(id casm.IDer) <-chan net.Stream {
	n.Lock()
	defer n.Unlock()
	return n.maybeInitUnsafe(id.ID())
}
