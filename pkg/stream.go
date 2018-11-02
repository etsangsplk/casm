package casm

import (
	"context"
	"io"
	gonet "net"
	"time"

	net "github.com/libp2p/go-libp2p-net"
)

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

type stream struct {
	c      context.Context
	cancel func()
	net.Stream
}

func newStream(c context.Context, s net.Stream) (str *stream) {
	str = new(stream)
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
	s.cancel()
	return s.Reset()
}

func (s stream) Read(b []byte) (n int, err error) {
	if n, err = s.Stream.Read(b); err != nil {
		s.maybePermanent(err)
	}
	return
}

func (s stream) Write(b []byte) (n int, err error) {
	if n, err = s.Stream.Write(b); err != nil {
		s.maybePermanent(err)
	}
	return
}

func (s stream) maybePermanent(err error) {
	if err, ok := err.(gonet.Error); ok && !err.Temporary() {
		s.cancel()
	}
}
