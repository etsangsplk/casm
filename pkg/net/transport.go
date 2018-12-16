package net

import (
	"context"

	pipe "github.com/lthibault/pipewerks/pkg"
	"github.com/lthibault/pipewerks/pkg/transport/generic"
	"github.com/lthibault/pipewerks/pkg/transport/inproc"
	"github.com/lthibault/pipewerks/pkg/transport/tcp"
	"github.com/pkg/errors"
)

// Transport is a means by which to connect to an listen for connections from
// other peers.
type Transport struct{ pipe.Transport }

// NewTransport from a pipewerks Transport
func NewTransport(a Addr) Transport {
	var pt pipe.Transport
	optMux := generic.OptMuxAdapter(connAdapter{})
	optNeg := generic.OptConnectHandler(idNegotiator(a.ID()))

	switch a.Network() {
	case "inproc":
		pt = inproc.New(
			inproc.OptDialback(a),
			inproc.OptGeneric(optMux),
			inproc.OptGeneric(optNeg),
		)
	case "tcp":
		pt = tcp.New(
			tcp.OptGeneric(optMux),
			tcp.OptGeneric(optNeg),
		)
	default:
		panic(errors.Errorf("invalid network %s", a.Network()))
	}

	return Transport{pt}
}

// Listen for connections
func (t Transport) Listen(c context.Context, a Addr) (Listener, error) {
	l, err := t.Transport.Listen(c, a)
	if err != nil {
		return Listener{}, err
	}

	// use the listener's address, because of address resolution. (e.g. ":80")
	a = NewAddr(a.ID(), l.Addr().Network(), l.Addr().String())
	return Listener{addr: a, Listener: l}, nil
}

// Dial into a remote listener
func (t Transport) Dial(c context.Context, a Addr) (*Conn, error) {
	conn, err := t.Transport.Dial(c, a)
	if err != nil {
		return nil, errors.Wrap(err, "transport")
	}

	return mkConn(conn.(pipeWrapper).idPair, conn), nil
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

	return mkConn(conn.(pipeWrapper).idPair, conn), nil
}
