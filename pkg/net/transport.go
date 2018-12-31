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
}

// ProtoListener can produce a listener that negotiates connection upgrades
// according to the casm network protocol.
type ProtoListener interface {
	Listen(context.Context) (*Listener, error)
}

// Upgrader negotiates a connection upgrade according to the CASM protocol
type Upgrader interface {
	BindDialback(Addr) func(Addr) DialUpgrader
	BindListen(Addr) ListenUpgrader
}

// DialUpgrader negotiates a connection upgrade from the dial endpoint.
type DialUpgrader interface {
	UpgradeDialer(pipe.Conn) error
}

// ListenUpgrader negotiates a connection upgrade from the listen endpoint.
type ListenUpgrader interface {
	UpgradeListener(pipe.Conn) (Addr, error)
}

type (
	pipeDialer interface {
		Dial(context.Context, net.Addr) (pipe.Conn, error)
	}

	pipeListener interface {
		Listen(context.Context, net.Addr) (pipe.Listener, error)
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
	return dialer{
		pipeDialer:      t.pt,
		local:           dialback,
		newDialUpgrader: t.u.BindDialback(dialback),
	}
}

// NewListener binds a listen Addr to the Transport
func (t Transport) NewListener(listen Addr) ProtoListener {
	return listener{
		pipeListener: t.pt,
		local:        listen,
		u:            t.u.BindListen(listen),
	}
}

type dialer struct {
	local Addr
	pipeDialer
	newDialUpgrader func(Addr) DialUpgrader
}

func (d dialer) Dial(c context.Context, a Addr) (*Conn, error) {
	du := d.newDialUpgrader(a)

	pc, err := d.pipeDialer.Dial(c, a)
	if err != nil {
		return nil, errors.Wrap(err, "dial pipe")
	}

	if err = du.UpgradeDialer(pc); err != nil {
		return nil, errors.Wrap(err, "upgrade")
	}

	return &Conn{Conn: pc, local: d.local, remote: a}, nil
}

type listener struct {
	local Addr
	pipeListener
	u ListenUpgrader
}

func (l listener) Listen(c context.Context) (*Listener, error) {
	pl, err := l.pipeListener.Listen(c, l.local)
	if err != nil {
		return nil, errors.Wrap(err, "listen pipe")
	}

	return &Listener{Listener: pl, a: l.local, u: l.u}, nil
}

// Listener can listen for incoming connections
type Listener struct {
	a Addr
	u ListenUpgrader
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

	a, err := l.u.UpgradeListener(conn)
	if err != nil {
		return nil, errors.Wrap(err, "upgrade")
	}

	return &Conn{Conn: conn, local: l.a, remote: a}, nil
}
