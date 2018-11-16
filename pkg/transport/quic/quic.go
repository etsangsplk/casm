package quic

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"io"
	"io/ioutil"

	"golang.org/x/sync/errgroup"

	"github.com/SentimensRG/ctx"
	net "github.com/lthibault/casm/pkg/net"
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

func (conn *conn) Negotiate(c context.Context, id net.IDer) (*conn, error) {
	conn.local = id.ID()

	s, err := conn.Session.OpenStream()
	if err != nil {
		return nil, errors.Wrap(err, "open stream")
	}
	defer s.Close()

	if t, ok := c.Deadline(); ok {
		if err = s.SetDeadline(t); err != nil {
			return nil, errors.Wrap(err, "set deadlines")
		}
	}

	var g errgroup.Group

	g.Go(func() error {
		ch := make(chan error, 1)
		go func() {
			if hdr, err := ioutil.ReadAll(io.LimitReader(s, 8)); err != nil {
				conn.remote = net.PeerID(binary.BigEndian.Uint64(hdr))
			}
			ch <- err
			close(ch)
		}()

		select {
		case err := <-ch:
			return errors.Wrap(err, "read header")
		case <-c.Done():
			return c.Err()
		}
	})

	g.Go(func() error {
		ch := make(chan error, 1)
		go func() {
			ch <- binary.Write(s, binary.BigEndian, id.ID())
			close(ch)
		}()

		select {
		case err := <-ch:
			return errors.Wrap(err, "write header")
		case <-c.Done():
			return c.Err()
		}
	})

	return conn, g.Wait()
}

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
	net.IDer
	q *Config
	t *tls.Config
}

// Dial the specified address
func (t *Transport) Dial(c context.Context, a net.Addr) (conn net.Conn, err error) {
	var sess quic.Session
	if sess, err = quic.DialAddrContext(c, a.String(), t.t, t.q); err != nil {
		err = errors.Wrap(err, "dial")
	} else if conn, err = mkConn(sess).Negotiate(c, a); err != nil {
		err = errors.Wrap(err, "negotiate")
	}

	return
}

// Listen on the specified address
func (t *Transport) Listen(c context.Context, a net.Addr) (net.Listener, error) {
	l, err := quic.ListenAddr(a.String(), t.t, t.q)
	if err != nil {
		return nil, err
	}
	ctx.Defer(c, func() { l.Close() })

	return &listener{Listener: l, IDer: t.ID()}, nil
}

type listener struct {
	net.IDer
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

		conn, err = mkConn(sess).Negotiate(c, l.ID())
		err = errors.Wrap(err, "negotiate")
	}

	return
}

// Option for Transport
type Option func(*Transport) (prev Option)

// OptQuic sets the QUIC configuration
func OptQuic(q *quic.Config) Option {
	return func(t *Transport) (prev Option) {
		prev = OptQuic(t.q)
		t.q = q
		return
	}
}

// OptTLS sets the TLS configuration
func OptTLS(tc *tls.Config) Option {
	return func(t *Transport) (prev Option) {
		prev = OptTLS(t.t)
		t.t = tc
		return
	}
}

// NewTransport over QUIC
func NewTransport(id net.IDer, opt ...Option) *Transport {
	t := &Transport{IDer: id.ID()}
	for _, o := range opt {
		o(t)
	}
	return t
}
