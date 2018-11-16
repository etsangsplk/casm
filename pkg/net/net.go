package net

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

// ErrorCode is used to terminate a connection and signal an error
type ErrorCode uint16

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

// Transport is a means by which to connect to an listen for connections from
// other peers.
type Transport interface {
	Listen(context.Context, Addr) (Listener, error)
	Dial(context.Context, Addr) (Conn, error)
}

// Listener can listen for incoming connections
type Listener interface {
	// Close the server
	Close() error
	// Accept returns new connections; this should be called in a loop.
	Accept(context.Context) (Conn, error)
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
	Path() string
	Context() context.Context
	Endpoint() EndpointPair
	io.Closer
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// PeerID is a unique identifier for a Node
type PeerID uint64

func (id PeerID) String() string { return fmt.Sprintf("%016x", uint64(id)) }

// ID satisfies the IDer interface
func (id PeerID) ID() PeerID { return id }

// Addr of a Host
type Addr interface {
	ID() PeerID
	Addr() Addr
	net.Addr
}

type addr struct {
	PeerID
	addr string
}

// NewAddr from an ID and an address stringer
func NewAddr(id PeerID, a string) Addr {
	return &addr{PeerID: id, addr: a}
}

func (a addr) Addr() Addr      { return a }
func (a addr) Network() string { return "udp" }
func (a addr) String() string  { return a.addr }
