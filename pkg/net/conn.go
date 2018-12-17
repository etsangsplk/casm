package net

import (
	"context"

	pipe "github.com/lthibault/pipewerks/pkg"
)

// Conn is a logical connection to a peer.  Streams are multiplexed onto Conns.
type Conn struct {
	local, remote Addr
	pipe.Conn
}

func mkConn(e edge, conn pipe.Conn) *Conn {
	return &Conn{
		local:  e.Local,
		remote: e.Remote,
		Conn:   conn,
	}
}

// LocalAddr of the connection
func (c Conn) LocalAddr() Addr { return c.local }

// RemoteAddr of the connection
func (c Conn) RemoteAddr() Addr { return c.remote }

// AcceptStream listens for the next incoming stream
func (c Conn) AcceptStream() (*Stream, error) {
	s, err := c.Conn.AcceptStream()
	return &Stream{Stream: s}, err
}

// OpenStream dials a stream
func (c Conn) OpenStream() (*Stream, error) {
	s, err := c.Conn.OpenStream()
	return &Stream{Stream: s}, err
}

// WithContext returns a new Stream, bound to the specified context.  Many
// applications assume Stream.Context() expires when the stream is closed, so
// use with care.
func (c Conn) WithContext(cx context.Context) *Conn {
	return &Conn{
		local:  c.local,
		remote: c.remote,
		Conn:   connCtxOverride{c: cx, Conn: c.Conn},
	}
}

type connCtxOverride struct {
	c context.Context
	pipe.Conn
}

func (o connCtxOverride) Context() context.Context { return o.c }
