package graph

import (
	"context"
	"io"
	"sync"

	"github.com/SentimensRG/ctx"
	"github.com/SentimensRG/ctx/mergectx"
	casm "github.com/lthibault/casm/pkg"
	"golang.org/x/sync/errgroup"
)

// streamGroup is a logical group of CASM streams.  They are all bound to the
// same root context and will all close if any one of them closes.
type streamGroup interface {
	Context() context.Context
	DataStream() casm.Stream
	CtrlStream() casm.Stream
	io.Closer
}

type idempotentCloser struct {
	sync.Once
	err error
	io.Closer
}

func (ic *idempotentCloser) Close() error {
	ic.Do(func() { ic.err = ic.Closer.Close() })
	return ic.err
}

type multiCloser []io.Closer

func newMultiCloser(close ...io.Closer) multiCloser {
	var mc multiCloser = make([]io.Closer, len(close))
	for i, c := range close {
		mc[i] = &idempotentCloser{Closer: c}
	}
	return mc
}

func (mc multiCloser) Close() error {
	var g errgroup.Group
	for _, c := range mc {
		g.Go(c.Close)
	}
	return g.Wait()
}

type syncStream struct {
	sync.RWMutex
	casm.Stream
}

func (s *syncStream) CloseWrite() (err error) {
	s.Lock()
	err = s.Stream.CloseWrite()
	s.Unlock()
	return
}

func (s *syncStream) Read(b []byte) (n int, err error) {
	s.RLock()
	n, err = s.Stream.Read(b)
	s.RUnlock()
	return
}

func (s *syncStream) Write(b []byte) (n int, err error) {
	s.RLock()
	n, err = s.Stream.Write(b)
	s.RUnlock()
	return
}

func (s *syncStream) Close() (err error) {
	s.Lock()
	err = s.Stream.Close()
	s.Unlock()
	return
}

type streamGrp struct {
	c          context.Context
	data, ctrl casm.Stream
	io.Closer
}

func newStreamGroup(data, ctrl casm.Stream) *streamGrp {
	c := mergectx.Link(data.Context(), ctrl.Context())
	c, cancel := context.WithCancel(c)

	synData := &syncStream{Stream: data}
	synCtrl := &syncStream{Stream: ctrl}

	ma := newMultiCloser(synData, synCtrl)
	ctx.Defer(c, func() {
		cancel()
		ma.Close()
	})

	return &streamGrp{
		c:      c,
		data:   synData,
		ctrl:   synCtrl,
		Closer: ma,
	}
}

func (g streamGrp) Context() context.Context { return g.c }
func (g streamGrp) DataStream() casm.Stream  { return g.data }
func (g streamGrp) CtrlStream() casm.Stream  { return g.ctrl }
