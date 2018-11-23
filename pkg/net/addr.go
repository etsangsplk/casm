package net

import (
	"context"
	"net"

	pipe "github.com/lthibault/pipewerks/pkg"
)

// Addr of a Host
type Addr interface {
	ID() PeerID
	Addr() Addr
	net.Addr
}

type addr struct {
	PeerID
	network, addr string
}

// NewAddr from an ID and an address stringer
func NewAddr(id PeerID, net, a string) Addr {
	return &addr{PeerID: id, network: net, addr: a}
}

func (a addr) Addr() Addr      { return a }
func (a addr) Network() string { return a.network }
func (a addr) String() string  { return a.addr }

type bind struct {
	c context.Context
	pipe.Stream
}

func (b bind) Context() context.Context { return b.c }

// Bind context to a stream
func Bind(c context.Context, s pipe.Stream) pipe.Stream {
	return bind{c: c, Stream: s}
}
