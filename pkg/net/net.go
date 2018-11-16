package net

import (
	"context"
	"io"
	"net"
	"time"
)

// ErrorCode is used to terminate a connection and signal an error
type ErrorCode uint16

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

// IDer can provide a PeerID
type IDer interface {
	ID() PeerID
}

// PeerID is a unique identifier for a Node
type PeerID uint64

// ID satisfies the IDer interface
func (id PeerID) ID() PeerID { return id }

// Addr of a Host
type Addr interface {
	IDer
	Addr() Addr
	net.Addr
}

type addr struct {
	IDer
	addr string
}

// NewAddr from an ID and an address stringer
func NewAddr(id IDer, a string) Addr {
	return &addr{IDer: id, addr: a}
}

func (a addr) Addr() Addr      { return a }
func (a addr) Network() string { return "udp" }
func (a addr) String() string  { return a.addr }
