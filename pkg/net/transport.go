package net

import (
	"context"
	"net"

	pipe "github.com/lthibault/pipewerks/pkg"
	"github.com/pkg/errors"
)

type (
	pipeDialer interface {
		Dial(context.Context, net.Addr) (pipe.Conn, error)
	}

	pipeListener interface {
		Listen(context.Context, net.Addr) (pipe.Listener, error)
	}

	dialUpgradeHandler interface {
		UpgradeDialer(pipe.Conn, Addr, PeerID) error
	}

	listenUpgradeHandler interface {
		UpgradeListener(pipe.Conn, Addr) (remote Addr, err error)
	}
)

// Transport is an abstraction over a reliable network connection.
type Transport struct{ pt pipe.Transport }

// NewTransport based on pipewerks.
func NewTransport(t pipe.Transport) *Transport {
	return &Transport{pt: t}
}

// NewDialer binds a dialback Addr to the Transport
func (t Transport) NewDialer(dialback Addr) ProtoDialer {
	return ProtoDialer{pipeDialer: t.pt, local: dialback, u: upgrader}
}

// NewListener binds a listen Addr to the Transport
func (t Transport) NewListener(listen Addr) ProtoListener {
	return ProtoListener{pipeListener: t.pt, local: listen, u: upgrader}
}

// ProtoDialer can initiate connection upgrades using the casm network protocol.
type ProtoDialer struct {
	local Addr
	pipeDialer
	u dialUpgradeHandler
}

// Dial into a remote Listener
func (d ProtoDialer) Dial(c context.Context, a Addr) (*Conn, error) {
	pc, err := d.pipeDialer.Dial(c, a)
	if err != nil {
		return nil, errors.Wrap(err, "dial pipe")
	}

	if err = d.u.UpgradeDialer(pc, d.local, a.ID()); err != nil {
		return nil, errors.Wrap(err, "upgrade")
	}

	return &Conn{Conn: pc, local: d.local, remote: a}, nil
}

// ProtoListener can produce a ProtoListener that negotiates connection upgrades
// according to the casm network protocol.
type ProtoListener struct {
	local Addr
	pipeListener
	u listenUpgradeHandler
}

// Listen for incoming connections
func (l ProtoListener) Listen(c context.Context) (*Listener, error) {
	pl, err := l.pipeListener.Listen(c, l.local)
	if err != nil {
		return nil, errors.Wrap(err, "listen pipe")
	}

	return &Listener{Listener: pl, a: l.local, u: l.u}, nil
}

// Listener can listen for incoming connections
type Listener struct {
	a Addr
	u listenUpgradeHandler
	pipe.Listener
}

// Addr is the local listen address
func (l Listener) Addr() Addr { return l.a }

// Accept the next incoming connection
func (l Listener) Accept() (*Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, errors.Wrap(err, "accept")
	}

	a, err := l.u.UpgradeListener(conn, l.a)
	if err != nil {
		return nil, errors.Wrap(err, "upgrade")
	}

	return &Conn{Conn: conn, local: l.a, remote: a}, nil
}
