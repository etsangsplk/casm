package net

import (
	"context"

	pipe "github.com/lthibault/pipewerks/pkg"
)

// Stream is a bidirectional connection between two hosts.
type Stream struct{ pipe.Stream }

// WithContext returns a new Stream, bound to the specified context.  Many
// applications assume Stream.Context() expires when the stream is closed, so
// use with care.
func (s Stream) WithContext(c context.Context) *Stream {
	return &Stream{Stream: streamCtxOverride{c: c, Stream: s.Stream}}
}

type streamCtxOverride struct {
	c context.Context
	pipe.Stream
}

func (o streamCtxOverride) Context() context.Context { return o.c }
