package net

import (
	"context"
	"net"

	pipe "github.com/lthibault/pipewerks/pkg"
	"github.com/pkg/errors"
)

// ProtoDialer can initiate connection upgrades using the casm network protocol.
type ProtoDialer interface {
	Dial(context.Context, Addr) (*Conn, error)
	UpgradeDialer(pipe.Conn, Addr, PeerID) error
}

// ProtoListener can produce a listener that negotiates connection upgrades
// according to the casm network protocol.
type ProtoListener interface {
	Listen(context.Context) (*Listener, error)
	UpgradeListener(pipe.Conn, Addr) (remote Addr, err error)
}

// Upgrader negotiates a connection upgrade according to the CASM protocol
type Upgrader interface {
	UpgradeDialer(pipe.Conn, Addr, PeerID) error
	UpgradeListener(pipe.Conn, Addr) (remote Addr, err error)
}

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

// Transport is an abstraction over a reliable network connection.  The NewDialer
// and NewListener methods bind a dialback/listen address, respectively, to the
// ProtoDialer & ProtoListener.
type Transport struct {
	pt pipe.Transport
	u  Upgrader
}

// NewTransport binds a pipewerks Transport to an Upgrader.
func NewTransport(t pipe.Transport, u Upgrader) *Transport {
	return &Transport{
		pt: t,
		u:  u,
	}
}

// NewDialer binds a dialback Addr to the Transport
func (t Transport) NewDialer(dialback Addr) ProtoDialer {
	return dialer{pipeDialer: t.pt, local: dialback, dialUpgradeHandler: t.u}
}

// NewListener binds a listen Addr to the Transport
func (t Transport) NewListener(listen Addr) ProtoListener {
	return listener{pipeListener: t.pt, local: listen, listenUpgradeHandler: t.u}
}

type dialer struct {
	local Addr
	pipeDialer
	dialUpgradeHandler
}

func (d dialer) Dial(c context.Context, a Addr) (*Conn, error) {
	pc, err := d.pipeDialer.Dial(c, a)
	if err != nil {
		return nil, errors.Wrap(err, "dial pipe")
	}

	// Handshake
	// 1. announce who we are to the listener (who has no idea right now)
	// 2. verify that PIDs match
	if err = d.UpgradeDialer(pc, d.local, a.ID()); err != nil {
		return nil, errors.Wrap(err, "upgrade")
	}

	return &Conn{Conn: pc, local: d.local, remote: a}, nil
}

type listener struct {
	local Addr
	pipeListener
	listenUpgradeHandler
}

func (l listener) Listen(c context.Context) (*Listener, error) {
	pl, err := l.pipeListener.Listen(c, l.local)
	if err != nil {
		return nil, errors.Wrap(err, "listen pipe")
	}

	return &Listener{Listener: pl, a: l.local, u: l}, nil
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

	// Handshake
	// 1. who called us?
	// 2. send our PID so that he can confirm that he reached the right entity
	a, err := l.u.UpgradeListener(conn, l.a)
	if err != nil {
		return nil, errors.Wrap(err, "upgrade")
	}

	return &Conn{Conn: conn, local: l.a, remote: a}, nil
}
