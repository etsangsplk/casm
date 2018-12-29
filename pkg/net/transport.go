package net

import (
	"context"

	pipe "github.com/lthibault/pipewerks/pkg"
	"github.com/pkg/errors"
)

// Transport is a means by which to connect to an listen for connections from
// other peers.
type Transport interface {
	Addr() Addr
	Listen(context.Context) (Listener, error)
	Dial(context.Context, Addr) (*Conn, error)
}

// TransportFactory initializes a net.Transport.  Listeneres produced by said
// Transport will bind to the specified address, and Dialers will transmit this
// same address as a dialback point.
type TransportFactory interface {
	NewTransport(Addr) (Transport, error)
}

// TransportFactoryFunc wraps an address in order to satisfy TransportFactory.
type TransportFactoryFunc func(Addr) (Transport, error)

// NewTransport satisfies TransportFactory
func (f TransportFactoryFunc) NewTransport(a Addr) (Transport, error) {
	return f(a)
}

// NewFactory builds a TransportFactory from a pipe.Transport
func NewFactory(t pipe.Transport) TransportFactory {
	return TransportFactoryFunc(func(a Addr) (Transport, error) {
		return pipeTransport{
			listen:    a,
			Transport: t,
		}, nil
	})
}

type pipeTransport struct {
	listen Addr
	pipe.Transport
}

func (t pipeTransport) Addr() Addr { return t.listen }

// Listen for connections
func (t pipeTransport) Listen(c context.Context) (Listener, error) {
	l, err := t.Transport.Listen(c, t.listen)
	if err != nil {
		return Listener{}, err
	}

	// use the listener's address, because of address resolution. (e.g. ":80")
	a := NewAddr(t.listen.ID(), l.Addr().Network(), l.Addr().String())
	return Listener{addr: a, Listener: l}, nil
}

// Dial into a remote listener
func (t pipeTransport) Dial(c context.Context, a Addr) (*Conn, error) {
	conn, err := t.Transport.Dial(c, a)
	if err != nil {
		return nil, errors.Wrap(err, "transport")
	}

	return mkConn(conn.(pipeWrapper).edge, conn), nil
}

// Listener can listen for incoming connections
type Listener struct {
	addr Addr
	pipe.Listener
}

// Addr is the local listen address
func (l Listener) Addr() Addr { return l.addr }

// Accept the next incoming connection
func (l Listener) Accept() (*Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, errors.Wrap(err, "accept")
	}

	return mkConn(conn.(pipeWrapper).edge, conn), nil
}
