package net

import (
	"context"
	"encoding/binary"

	pipe "github.com/lthibault/pipewerks/pkg"
	"github.com/pkg/errors"
)

// Transport is a means by which to connect to an listen for connections from
// other peers.
type Transport struct{ pipe.Transport }

// Listen for connections
func (t Transport) Listen(c context.Context, a Addr) (Listener, error) {
	l, err := t.Transport.Listen(c, a)
	return Listener{PeerID: a.ID(), Listener: l}, err
}

// Dial into a remote listener
func (t Transport) Dial(c context.Context, local PeerID, a Addr) (*Conn, error) {
	conn, err := t.Transport.Dial(c, a)
	if err != nil {
		return nil, errors.Wrap(err, "transport")
	}

	raw := conn.(pipe.RawConn).Raw()

	// send local ID to the peer
	if err = binary.Write(raw, binary.BigEndian, local); err != nil {
		raw.Close()
		return nil, errors.Wrap(err, "handshake")
	}

	return &Conn{localID: local, remoteID: a.Addr().ID(), Conn: conn}, nil
}

// Listener can listen for incoming connections
type Listener struct {
	PeerID
	pipe.Listener
}

// Addr is the local listen address
func (l Listener) Addr() Addr {
	return NewAddr(l.PeerID, l.Addr().Network(), l.Addr().String())
}

// Accept the next incoming connection
func (l Listener) Accept(c context.Context) (*Conn, error) {
	conn, err := l.Listener.Accept(c)
	if err != nil {
		return nil, errors.Wrap(err, "accept")
	}

	// get the remote ID
	var remote PeerID
	raw := conn.(pipe.RawConn).Raw()
	if err = binary.Read(raw, binary.BigEndian, &remote); err != nil {
		return nil, errors.Wrap(err, "handshake")
	}

	return &Conn{localID: l.Addr().ID(), remoteID: remote, Conn: conn}, nil
}

// Conn is a logical connection to a peer.  Streams are multiplexed onto Conns.
type Conn struct {
	localID, remoteID PeerID
	pipe.Conn
}

// Endpoint provides address information
func (c Conn) Endpoint() Edge {
	ep := c.Endpoint()
	local := ep.Local()
	remote := ep.Remote()
	return Edge{
		local:  NewAddr(c.localID, local.Network(), local.String()),
		remote: NewAddr(c.remoteID, remote.Network(), remote.String()),
	}
}

// Edge provides the endpoints of a connection
type Edge struct{ local, remote Addr }

// Local peer address
func (e Edge) Local() Addr { return e.local }

// Remote peer address
func (e Edge) Remote() Addr { return e.remote }

// Handler responds to an incoming stream connection
type Handler interface {
	Serve(pipe.Stream)
}

// HandlerFunc is an adapter to allow the use of ordinary functions as stream
// handlers.  If f is a function with the appropriate signature, HandlerFunc(f)
// is a Handler that calls f.
type HandlerFunc func(pipe.Stream)

// Serve satisfies Handler.  It calls h.
func (h HandlerFunc) Serve(s pipe.Stream) { h(s) }
