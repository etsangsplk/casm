package net

import (
	"io"
	"net"
	"time"

	pipe "github.com/lthibault/pipewerks/pkg"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	upgradeDeadline = time.Second * 5
)

var (
	proto    protocol
	upgrader pipeConnUpgrader
)

type pipeConnUpgrader struct{}

// UpgradeDialer satisfies Upgrader
func (u pipeConnUpgrader) UpgradeDialer(conn pipe.Conn, local Addr, remote PeerID) error {
	s, err := conn.OpenStream()
	if err != nil {
		return errors.Wrap(err, "open stream")
	}

	return proto.upgradeDialer(s, local, remote)
}

// UpgradeListener satisfies Upgrader
func (u pipeConnUpgrader) UpgradeListener(conn pipe.Conn, local Addr) (remote Addr, err error) {
	s, err := conn.AcceptStream()
	if err != nil {
		return nil, errors.Wrap(err, "accept stream")
	}

	return proto.upgradeListener(s, local)
}

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
		if err = binary.Read(r, &remote); err != nil {
			err = errors.Wrap(err, "read remote ID")
		} else if remote != id {
			err = errors.Errorf("expected remote peer %s, got %s", id, remote)
		}
		return
	}
}

func sendDialback(w io.Writer, a Addr) func() error {
	return func() error {
		return errors.Wrap(newWireAddr(a).SendTo(w), "send dialback")
	}
}

func sendID(w io.Writer, id PeerID) func() error {
	return func() (err error) {
		if err = binary.Write(w, id); err != nil {
			return errors.Wrap(err, "transmit local ID")
		}
		return
	}
}

func recvDialback(r io.Reader, a *wireAddr) func() error {
	return func() error {
		return errors.Wrap(a.RecvFrom(r), "recv dialback")
	}
}
