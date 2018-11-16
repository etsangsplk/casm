package graph

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/SentimensRG/ctx/mergectx"
	net "github.com/lthibault/casm/pkg/net"
	"golang.org/x/sync/errgroup"
)

// streamGroup is a logical group of CASM streams.  They are all bound to the
// same root context and will all close if any one of them closes.
type streamGroup interface {
	Context() context.Context
	DataStream() net.Stream
	CtrlStream() net.Stream
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
	net.Stream
	c      context.Context
	cancel func()
}

func (s *syncStream) Context() context.Context { return s.c }

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
	s.cancel()
	s.Lock()
	err = s.Stream.Close()
	s.Unlock()
	return
}

func (s *syncStream) SetDeadline(t time.Time) (err error) {
	s.RLock()
	err = s.Stream.SetDeadline(t)
	s.RUnlock()
	return
}

func (s *syncStream) SetReadDeadline(t time.Time) (err error) {
	s.RLock()
	err = s.Stream.SetReadDeadline(t)
	s.RUnlock()
	return
}

func (s *syncStream) SetWriteDeadline(t time.Time) (err error) {
	s.RLock()
	err = s.Stream.SetWriteDeadline(t)
	s.RUnlock()
	return
}

type streamGrp struct {
	c          context.Context
	data, ctrl net.Stream
	io.Closer
}

func newStreamGroup(data, ctrl net.Stream) *streamGrp {
	c := mergectx.Link(data.Context(), ctrl.Context())
	c, cancel := context.WithCancel(c)

	synData := &syncStream{Stream: data, c: c, cancel: cancel}
	synCtrl := &syncStream{Stream: ctrl, c: c, cancel: cancel}

	return &streamGrp{
		c:      c,
		data:   synData,
		ctrl:   synCtrl,
		Closer: newMultiCloser(synData, synCtrl),
	}
}

func (g streamGrp) Context() context.Context { return g.c }
func (g streamGrp) DataStream() net.Stream   { return g.data }
func (g streamGrp) CtrlStream() net.Stream   { return g.ctrl }
