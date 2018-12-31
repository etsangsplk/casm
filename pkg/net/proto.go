package net

import (
	"encoding/binary"
	"io"
	"net"
	"time"

	"github.com/lunixbochs/struc"

	pipe "github.com/lthibault/pipewerks/pkg"
	"github.com/lthibault/pipewerks/pkg/transport/generic"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	upgradeDeadline = time.Second * 5
)

// RawConnUpgrader can negotiate a connection upgrade for any pipewerks transport
// that features raw connection hooks.  This is the case for all transports built
// on top of pipewerks' `generic` transport, e.g. TCP.
//
// N.B.: RawConnUpgrader must be passed both to NewTransport as an Upgrader as
// well as to the pipe.Transport constructor via a `generic.OptConnectHandler`.
type RawConnUpgrader struct{}

// Connected satisfies pipewerks' generic.OnConnect
func (u RawConnUpgrader) Connected(conn net.Conn, e generic.EndpointType) (net.Conn, error) {
	switch e {
	case generic.DialEndpoint:
		panic("NOT IMPLEMENTED")
	case generic.ListenEndpoint:
		panic("NOT IMPLEMENTED")
	default:
		panic("unreachable")
	}
}

// UpgradeDialer satisfies Upgrader
func (u RawConnUpgrader) UpgradeDialer(conn pipe.Conn, remote PeerID) error {
	panic("function NOT IMPLEMENTED")
}

// UpgradeListener satisfies Upgrader
func (u RawConnUpgrader) UpgradeListener(conn pipe.Conn, local Addr) (remote Addr, err error) {
	panic("function NOT IMPLEMENTED")
}

// PipeConnUpgrader uses pipewerks Streams to negotiate the connection upgrade.
// It is compatible with any pipe.Transport.
//
// N.B.: PipeConnUpgrader must open and close a stream in order for negotiation
// to take place, which may increase latency.
type PipeConnUpgrader struct{}

// BindDialback address to a partial DialUpgrader
func (PipeConnUpgrader) BindDialback(a Addr) func(Addr) DialUpgrader {
	return func(remote Addr) DialUpgrader {
		return pcDialUpgrader{
			local:  a,
			remote: remote,
		}
	}
}

// BindListen address to a ListenUpgrader
func (PipeConnUpgrader) BindListen(a Addr) ListenUpgrader {
	return pcListenUpgrader{local: a}
}

type pcDialUpgrader struct{ local, remote Addr }

func (u pcDialUpgrader) UpgradeDialer(conn pipe.Conn) error {
	s, err := conn.OpenStream()
	if err != nil {
		return errors.Wrap(err, "open stream")
	}
	defer s.Close()

	return proto.upgradeDialer(s, u.local, u.remote.ID())
}

type pcListenUpgrader struct{ local Addr }

func (u pcListenUpgrader) UpgradeListener(conn pipe.Conn) (remote Addr, err error) {
	s, err := conn.AcceptStream()
	if err != nil {
		return nil, errors.Wrap(err, "accept stream")
	}
	defer s.Close()

	return proto.upgradeListener(s, u.local)
}

var proto protocol

type protocol struct{}

func withTimeout(conn net.Conn, fn func() error) error {
	err := conn.SetDeadline(time.Now().Add(upgradeDeadline))
	if err != nil {
		return errors.Wrap(err, "set deadline")
	}
	defer conn.SetDeadline(time.Time{})

	return fn()
}

func (p protocol) upgradeDialer(conn net.Conn, local Addr, remote PeerID) error {
	return withTimeout(conn, func() error {
		var g errgroup.Group
		g.Go(checkRemoteID(conn, remote))
		g.Go(sendDialback(conn, local))
		return g.Wait()
	})
}

func (protocol) upgradeListener(conn net.Conn, local Addr) (Addr, error) {
	a := new(wireAddr)
	return a, withTimeout(conn, func() error {
		var g errgroup.Group
		g.Go(sendID(conn, local.ID()))
		g.Go(recvDialback(conn, a))
		return g.Wait()
	})
}

func checkRemoteID(r io.Reader, id PeerID) func() error {
	var remote PeerID
	return func() (err error) {
		if err = binary.Read(r, binary.BigEndian, &remote); err != nil {
			err = errors.Wrap(err, "read remote ID")
		} else if remote != id {
			err = errors.Errorf("expected remote peer %s, got %s", id, remote)
		}
		return
	}
}

func sendDialback(w io.Writer, a Addr) func() error {
	return func() error {
		return errors.Wrap(struc.Pack(w, newWireAddr(a)), "send dialback")
	}
}

func sendID(w io.Writer, id PeerID) func() error {
	return func() (err error) {
		if err = binary.Write(w, binary.BigEndian, id); err != nil {
			return errors.Wrap(err, "transmit local ID")
		}
		return
	}
}

func recvDialback(r io.Reader, a *wireAddr) func() error {
	return func() error {
		return errors.Wrap(struc.Unpack(r, a), "recv dialback")
	}
}
