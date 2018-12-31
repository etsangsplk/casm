package net

import (
	"context"

	pipe "github.com/lthibault/pipewerks/pkg"
)

type localRemoteAddresser interface {
	LocalAddr() Addr
	RemoteAddr() Addr
}

// Stream is a bidirectional connection between two hosts.
type Stream struct {
	pipe.Stream
	addrs localRemoteAddresser
}

// LocalAddr of the stream
func (s *Stream) LocalAddr() Addr { return s.addrs.LocalAddr() }

// RemoteAddr of the peer
func (s *Stream) RemoteAddr() Addr { return s.addrs.RemoteAddr() }

// WithContext returns a new Stream, bound to the specified context.  Many
// applications assume Stream.Context() expires when the stream is closed, so
// use with care.
func (s *Stream) WithContext(c context.Context) *Stream {
	s.Stream = streamCtxOverride{c: c, Stream: s.Stream}
	return s
}

type streamCtxOverride struct {
	c context.Context
	pipe.Stream
}

func (o streamCtxOverride) Context() context.Context { return o.c }
