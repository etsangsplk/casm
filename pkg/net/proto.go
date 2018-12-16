package net

import (
	"encoding/binary"
	"net"
	"time"

	pipe "github.com/lthibault/pipewerks/pkg"
	"github.com/lthibault/pipewerks/pkg/transport/generic"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	handshakeTimeout = time.Second * 5
)

type idNegotiator PeerID

func (n idNegotiator) Connected(conn net.Conn, _ generic.EndpointType) (net.Conn, error) {
	err := conn.SetDeadline(time.Now().Add(handshakeTimeout))
	if err != nil {
		return nil, errors.Wrap(err, "set deadline")
	}

	var g errgroup.Group
	g.Go(func() error {
		return errors.Wrap(
			binary.Write(conn, binary.BigEndian, n),
			"write",
		)
	})

	var id PeerID
	g.Go(func() error {
		return errors.Wrap(
			binary.Read(conn, binary.BigEndian, &id),
			"read",
		)
	})

	if err = g.Wait(); err != nil {
		return nil, errors.Wrap(err, "handshake")
	}

	if err = conn.SetDeadline(time.Time{}); err != nil {
		return nil, errors.Wrap(err, "disable deadline")
	}

	return netWrapper{
		Conn:   conn,
		idPair: idPair{Local: PeerID(n), Remote: id},
	}, nil
}

type idPair struct{ Local, Remote PeerID }

type netWrapper struct {
	net.Conn
	idPair
}

type pipeWrapper struct {
	pipe.Conn
	idPair
}

// connAdapter wraps pipewerks' generic.MuxConfig to supply PeerIDs to the
// Transport Dial/Listen functions.
type connAdapter struct{ generic.MuxConfig }

func (a connAdapter) adapt(f func(net.Conn) (pipe.Conn, error), conn net.Conn) (pipe.Conn, error) {
	id := conn.(netWrapper)
	pc, err := f(conn)
	return pipeWrapper{
		Conn:   pc,
		idPair: id.idPair,
	}, err
}

func (a connAdapter) AdaptServer(conn net.Conn) (pipe.Conn, error) {
	return a.adapt(a.MuxConfig.AdaptServer, conn)
}

func (a connAdapter) AdaptClient(conn net.Conn) (pipe.Conn, error) {
	return a.adapt(a.MuxConfig.AdaptClient, conn)
}
