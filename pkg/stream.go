package casm

import (
	"context"
	"io"
	"sync"
	"time"
	"unsafe"

	net "github.com/libp2p/go-libp2p-net"
)

var strmPool = streamPool(sync.Pool{New: func() interface{} {
	return new(stream)
}})

// Stream is a bidirectional connection between two hosts.  Callers MUST call
// Close before discarding the stream.
type Stream interface {
	// Context expires when the connection is closed
	Context() context.Context
	CloseWrite() error
	io.ReadWriteCloser
	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// Handler responds to an incoming stream connection
type Handler interface {
	ServeStream(Stream)
}

// HandlerFunc is an adapter to allow the use of ordinary functions as stream
// handlers.  If f is a function with the appropriate signature, HandlerFunc(f)
// is a Handler that calls f.
type HandlerFunc func(Stream)

// ServeStream satisfies Handler.  It calls h.
func (h HandlerFunc) ServeStream(s Stream) { h(s) }

type streamPool sync.Pool

func (p *streamPool) Get() *stream {
	return (*sync.Pool)(unsafe.Pointer(p)).Get().(*stream)
}

func (p *streamPool) Put(s *stream) { (*sync.Pool)(unsafe.Pointer(p)).Put(s) }

type stream struct {
	c      context.Context
	cancel func()
	net.Stream
}

func newStream(c context.Context, s net.Stream) (str *stream) {
	str = strmPool.Get()
	str.c, str.cancel = context.WithCancel(c)
	str.Stream = s
	return
}

func (s stream) Context() context.Context { return s.c }

// CloseWrite clsoes the stream for writing.  Reading will still work (i.e.: the
// remote side can still write).
func (s stream) CloseWrite() error { return s.Stream.Close() }

// Close *MUST* be called before discarding the stream
func (s stream) Close() error {
	defer s.free()
	return s.Reset()
}

func (s *stream) free() {
	s.cancel()
	strmPool.Put(s)
}
