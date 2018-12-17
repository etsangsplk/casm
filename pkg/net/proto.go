package net

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"time"

	"golang.org/x/sync/errgroup"

	pipe "github.com/lthibault/pipewerks/pkg"
	"github.com/lthibault/pipewerks/pkg/transport/generic"
	"github.com/pkg/errors"
)

const (
	handshakeTimeout = time.Second * 5

	transInproc trans = iota
	transTCP
)

type trans uint8

type leadingHeader struct {
	PeerID
	Trans trans
	Len   uint8
}

func (l leadingHeader) addrReader(r io.Reader) io.Reader {
	return io.LimitReader(r, int64(l.Len))
}

func (l leadingHeader) Network() string {
	switch l.Trans {
	case transInproc:
		return "inproc"
	case transTCP:
		return "tcp"
	default:
		panic("invalid transport type")
	}
}

func newHdrParts(a Addr) (leadingHeader, string) {
	var l leadingHeader
	l.PeerID = a.ID()
	l.Len = uint8(len(a.String()))

	switch a.Network() {
	case "inproc":
		l.Trans = transInproc
	case "tcp":
		l.Trans = transTCP
	default:
		panic("unrecognized transport " + a.Network())
	}

	return l, a.String()
}

type negotiator interface {
	SendHdr(io.Writer) func() error
	RecvHdr(io.Reader) func() error
	Addr() Addr
}

type baseNegotiator chan Addr

func (b baseNegotiator) Addr() Addr { return <-b }

type dialNegotiator struct {
	baseNegotiator
	Local  Addr
	Remote net.Addr
}

func negotiateDial(local Addr, remote net.Addr) dialNegotiator {
	return dialNegotiator{
		baseNegotiator: make(chan Addr, 1),
		Local:          local,
		Remote:         remote,
	}
}

func (d dialNegotiator) SendHdr(w io.Writer) func() error {
	b := new(bytes.Buffer)
	head, tail := newHdrParts(d.Local)

	binary.Write(b, binary.BigEndian, head)
	b.WriteString(tail)

	return func() (err error) {
		_, err = io.Copy(w, b)
		err = errors.Wrap(err, "send full hdr")
		return
	}
}

func (d dialNegotiator) RecvHdr(r io.Reader) func() error {
	var pid PeerID
	return func() (err error) {
		if err = binary.Read(r, binary.BigEndian, &pid); err == nil {
			d.baseNegotiator <- addr{
				PeerID:  pid,
				network: d.Remote.Network(),
				addr:    d.Remote.String(),
			}
		}

		err = errors.Wrap(err, "recv partial hdr")
		close(d.baseNegotiator)

		return
	}
}

type listenNegotiator struct {
	baseNegotiator
	Local Addr
}

func negotiateListen(a Addr) listenNegotiator {
	return listenNegotiator{
		baseNegotiator: make(chan Addr),
		Local:          a,
	}
}

func (l listenNegotiator) SendHdr(w io.Writer) func() error {
	return func() error {
		return errors.Wrap(
			binary.Write(w, binary.BigEndian, l.Local.ID()),
			"send partial hdr",
		)
	}
}

func (l listenNegotiator) RecvHdr(r io.Reader) func() error {
	var h leadingHeader
	return func() (err error) {
		if err = binary.Read(r, binary.BigEndian, &h); err != nil {
			return errors.Wrap(err, "read full header (leading)")
		}

		b := new(bytes.Buffer)
		if _, err = io.Copy(b, h.addrReader(r)); err != nil {
			return errors.Wrap(err, "read full header (trailing)")
		}

		l.baseNegotiator <- addr{
			PeerID:  h.PeerID,
			network: h.Network(),
			addr:    b.String(),
		}
		close(l.baseNegotiator)

		return
	}
}

type handshakeProtocol struct{ a Addr }

func (h handshakeProtocol) Connected(conn net.Conn, e generic.EndpointType) (net.Conn, error) {
	err := conn.SetDeadline(time.Now().Add(handshakeTimeout))
	if err != nil {
		return nil, errors.Wrap(err, "set deadline")
	}

	var remote Addr
	switch e {
	case generic.DialEndpoint:
		// the listener has no idea who we are
		remote, err = h.negotiate(negotiateDial(h.a, conn.RemoteAddr()), conn)
	case generic.ListenEndpoint:
		// the dialer only needs our ID
		remote, err = h.negotiate(negotiateListen(h.a), conn)
	}

	if err != nil {
		return nil, err
	}

	if err = conn.SetDeadline(time.Time{}); err != nil {
		return nil, errors.Wrap(err, "disable deadline")
	}

	return netWrapper{
		Conn: conn,
		edge: edge{Local: h.a, Remote: remote},
	}, nil
}

func (h handshakeProtocol) negotiate(n negotiator, rw io.ReadWriter) (Addr, error) {
	var g errgroup.Group
	g.Go(n.SendHdr(rw))
	g.Go(n.RecvHdr(rw))
	return n.Addr(), g.Wait()
}

type edge struct{ Local, Remote Addr }

type netWrapper struct {
	net.Conn
	edge
}

type pipeWrapper struct {
	pipe.Conn
	edge
}

// connAdapter wraps pipewerks' generic.MuxConfig to supply PeerIDs to the
// Transport Dial/Listen functions.
type connAdapter struct{ generic.MuxConfig }

func (a connAdapter) adapt(f func(net.Conn) (pipe.Conn, error), conn net.Conn) (pipe.Conn, error) {
	e := conn.(netWrapper)
	pc, err := f(conn)
	return pipeWrapper{
		Conn: pc,
		edge: e.edge,
	}, err
}

func (a connAdapter) AdaptServer(conn net.Conn) (pipe.Conn, error) {
	return a.adapt(a.MuxConfig.AdaptServer, conn)
}

func (a connAdapter) AdaptClient(conn net.Conn) (pipe.Conn, error) {
	return a.adapt(a.MuxConfig.AdaptClient, conn)
}
