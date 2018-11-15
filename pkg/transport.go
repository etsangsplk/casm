package casm

import (
	"context"
	"io"
	"time"
)

// ErrorCode is used to terminate a connection and signal an error
type ErrorCode uint16

// Listener can listen for incoming connections
type Listener interface {
	// Close the server
	Close() error
	// Addr returns the local network addr on which the server is listening
	Addr() Addr
	// Accept returns new connections; this should be called in a loop.
	Accept(context.Context) (Conn, error)
}

// Transport is a means by which to connect to an listen for connections from
// other peers.
type Transport interface {
	Dial(context.Context, Addr) (Conn, error) // NOTE: use quic.DialAddrContext
	Listen(context.Context, Addr) (Listener, error)
}

// Conn represents a logical connection between two peers.  Streams are
// multiplexed onto connections
type Conn interface {
	Context() context.Context
	Stream() Streamer
	Endpoint() EndpointPair
	io.Closer
	CloseWithError(ErrorCode, error) error
}

// Streamer can open and close various types of streams
type Streamer interface {
	Accept() (Stream, error)
	Open() (Stream, error)
}

// EndpointPair identifies the two endpoints
type EndpointPair interface {
	Local() Addr
	Remote() Addr
}

// Stream is a bidirectional connection between two hosts
type Stream interface {
	Context() context.Context
	Endpoint() EndpointPair
	io.Closer
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// Dial into a transport
func Dial(c context.Context, t Transport, a Addr) (Conn, error) {
	return t.Dial(context.WithValue(c, keyListenAddr, a), a)
}
