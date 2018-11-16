package casm

import (
	"context"
	"encoding/binary"

	net "github.com/lthibault/casm/pkg/net"
)

// Listener can listen for incoming connections
type Listener interface {
	// Close the server
	Close() error
	// Addr returns the local network addr on which the server is listening
	Addresser
	// Accept returns new connections; this should be called in a loop.
	Accept(context.Context) (Conn, error)
}

// Transport is a means by which to connect to an listen for connections from
// other peers.
type Transport interface {
	Dial(context.Context, net.Addr) (Conn, error) // NOTE: use quic.DialAddrContext
	Listen(context.Context, net.Addr) (Listener, error)
}

// Dial into a transport
func Dial(c context.Context, t Transport, a Addr) (Conn, error) {
	header := make([]byte, 8) // consider sync.Pool
	binary.BigEndian.PutUint64(header, uint64(a.ID()))
	return t.Dial(context.WithValue(c, keyListenAddr, a), a)
}
