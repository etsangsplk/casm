package quic

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"io"
	"io/ioutil"

	"github.com/SentimensRG/ctx"
	net "github.com/lthibault/casm/pkg/net"
	log "github.com/lthibault/log/pkg"
	quic "github.com/lucas-clemente/quic-go"
	"github.com/pkg/errors"
)

// Config for QUIC protocol
type Config = quic.Config

type conn struct {
	local, remote net.PeerID
	quic.Session
}

func mkConn(s quic.Session) *conn {
	return &conn{Session: s}
}

func (conn *conn) SetLocalID(id net.PeerID)   { conn.local = id }
func (conn *conn) SetRemoteID(id net.PeerID)  { conn.remote = id }
func (conn *conn) Stream() net.Streamer       { return conn }
func (conn conn) Accept() (net.Stream, error) { return conn.Accept() }

func (conn conn) Open() (net.Stream, error) {
	s, err := conn.OpenStream()
	if err != nil {
		return nil, errors.Wrap(err, "open stream")
	}

	var size uint16
	if err = binary.Read(s, binary.BigEndian, &size); err != nil {
		return nil, errors.Wrap(err, "read pathsize")
	}

	path, err := ioutil.ReadAll(io.LimitReader(s, int64(size)))
	if err != nil {
		return nil, errors.Wrap(err, "read path")
	}

	return &stream{
		path:         string(path),
		Stream:       s,
		EndpointPair: conn,
	}, nil
}

func (conn *conn) Endpoint() net.EndpointPair { return conn }

func (conn conn) Local() net.Addr {
	return net.NewAddr(conn.local, conn.LocalAddr().String())
}

func (conn conn) Remote() net.Addr {
	return net.NewAddr(conn.remote, conn.RemoteAddr().String())
}

func (conn conn) CloseWithError(c net.ErrorCode, err error) error {
	return conn.Session.CloseWithError(quic.ErrorCode(c), err)
}

type stream struct {
	path string
	quic.Stream
	net.EndpointPair
}

func (s stream) Path() string               { return s.path }
func (s stream) Endpoint() net.EndpointPair { return s.EndpointPair }

// Transport over QUIC
type Transport struct {
	net.PeerID
	q *Config
	t *tls.Config
}

// Dial the specified address
func (t *Transport) Dial(c context.Context, a net.Addr) (net.Conn, error) {
	log.Get(c).Debug("dialing")

	sess, err := quic.DialAddrContext(c, a.String(), t.t, t.q)
	if err != nil {
		return nil, errors.Wrap(err, "dial")
	}

	log.Get(c).Debug("negotiating")

	conn := mkConn(sess)
	if err := net.NegotiateConn(c, a.ID(), conn); err != nil {
		return nil, errors.Wrap(err, "negotiate")
	}

	return conn, nil
}

// Listen on the specified address
func (t *Transport) Listen(c context.Context, a net.Addr) (net.Listener, error) {
	log.Get(c).Debug("listening")

	l, err := quic.ListenAddr(a.String(), t.t, t.q)
	if err != nil {
		return nil, err
	}
	ctx.Defer(c, func() { l.Close() })

	return &listener{Listener: l, PeerID: t.PeerID}, nil
}

type listener struct {
	net.PeerID
	quic.Listener
}

func (l listener) Accept(c context.Context) (conn net.Conn, err error) {
	cherr := make(chan error) // TODO:  sync.Pool

	var sess quic.Session

	go func() {
		var e error
		sess, e = l.Listener.Accept()

		select {
		case <-c.Done():
		case cherr <- errors.Wrap(e, "accept quic"):
		}
	}()

	select {
	case <-c.Done():
		err = c.Err()
	case err = <-cherr:
		if err != nil {
			return
		}

		rawConn := mkConn(sess)
		if err = net.NegotiateConn(c, l.ID(), rawConn); err != nil {
			err = errors.Wrap(err, "negotiate")
		} else {
			conn = rawConn
		}
	}

	return
}

// New Transport over QUIC
func New(id net.PeerID, opt ...Option) *Transport {
	t := &Transport{PeerID: id}
	for _, o := range opt {
		o(t)
	}
	return t
}
