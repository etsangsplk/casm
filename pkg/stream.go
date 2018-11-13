package casm

import (
	"context"
	"io"
	gonet "net"
	"time"

	net "github.com/libp2p/go-libp2p-net"
)

type contexter interface {
	Context() context.Context
}

// Stream is a bidirectional connection between two hosts.  Callers MUST call
// Close before discarding the stream.
type Stream interface {

	// Context expires when the connection is closed
	Context() context.Context
	RemotePeer() PeerID

	// CloseWrite closes the stream for writing.  Reading will still work (i.e.:
	// the remote side can still write).
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
	c context.Context
	IDer
	cancel func()
	net.Stream
}

func newStream(c context.Context, id IDer, s net.Stream) (str *stream) {
	str = new(stream)
	if cxr, ok := s.(contexter); ok {
		str.c = cxr.Context()
	} else {
		str.c, str.cancel = context.WithCancel(c)
	}
	str.IDer = id
	str.Stream = s
	return
}

func (s stream) Context() context.Context {
	if c, ok := s.Stream.(contexter); ok {
		return c.Context()
	}

	return s.c
}

func (s stream) RemotePeer() PeerID { return s.ID() }

// CloseWrite closes the stream for writing.  Reading will still work (i.e.: the
// remote side can still write).
func (s stream) CloseWrite() error { return s.Stream.Close() }

// Close *MUST* be called before discarding the stream
func (s stream) Close() error {
	if _, ok := s.Stream.(contexter); !ok {
		s.cancel()
	}
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
	if _, ok := s.Stream.(contexter); ok {
		if err, ok := err.(gonet.Error); ok && !err.Temporary() {
			s.cancel()
		}
	}
}

func (s stream) SetDeadline(t time.Time) (err error) {
	if err = s.Stream.SetDeadline(t); err != nil {
		s.maybePermanent(err)
	}
	return
}

func (s stream) SetReadDeadline(t time.Time) (err error) {
	if err = s.Stream.SetReadDeadline(t); err != nil {
		s.maybePermanent(err)
	}
	return
}

func (s stream) SetWriteDeadline(t time.Time) (err error) {
	if err = s.Stream.SetWriteDeadline(t); err != nil {
		s.maybePermanent(err)
	}
	return
}
