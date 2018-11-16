package net

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// NewAddr from an ID and an address stringer
func NewAddr(id PeerID, a string) Addr {
	return &addr{PeerID: id, addr: a}
}

func (a addr) Addr() Addr      { return a }
func (a addr) Network() string { return "udp" }
func (a addr) String() string  { return a.addr }

// ErrorCode is used to terminate a connection and signal an error
type ErrorCode uint16

// Handler responds to an incoming stream connection
type Handler interface {
	Serve(Stream)
}

// HandlerFunc is an adapter to allow the use of ordinary functions as stream
// handlers.  If f is a function with the appropriate signature, HandlerFunc(f)
// is a Handler that calls f.
type HandlerFunc func(Stream)

// Serve satisfies Handler.  It calls h.
func (h HandlerFunc) Serve(s Stream) { h(s) }

// Transport is a means by which to connect to an listen for connections from
// other peers.
type Transport interface {
	Listen(context.Context, Addr) (Listener, error)
	Dial(context.Context, Addr) (Conn, error)
}

// Listener can listen for incoming connections
type Listener interface {
	// Close the server
	Close() error
	Addr() net.Addr
	// Accept returns new connections; this should be called in a loop.
	Accept(context.Context) (Conn, error)
}

// Conn represents a logical connection between two peers.  Streams are
// multiplexed onto connections
type Conn interface {
	Context() context.Context
	Stream() Streamer
	Endpoint() EndpointPair
	io.Closer
	CloseWithError(ErrorCode, error) error
}

// RawConn is a connection for which protocol negotiation has not yet
// occurred
type RawConn interface {
	Conn
	SetLocalID(PeerID)
	SetRemoteID(PeerID)
}

// Streamer can open and close various types of streams
type Streamer interface {
	Accept() (Stream, error)
	Open() (Stream, error)
}

// EndpointPair identifies the two endpoints
type EndpointPair interface {
	Local() Addr
	Remote() Addr
}

// Stream is a bidirectional connection between two hosts
type Stream interface {
	Path() string
	Context() context.Context
	Endpoint() EndpointPair
	io.Closer
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// Addr of a Host
type Addr interface {
	ID() PeerID
	Addr() Addr
	net.Addr
}

type addr struct {
	PeerID
	addr string
}

// NegotiateConn handles protocol negotiation
func NegotiateConn(c context.Context, local PeerID, conn RawConn) error {
	s, err := conn.Stream().Open()
	if err != nil {
		return errors.Wrap(err, "open stream")
	}
	defer s.Close()

	if t, ok := c.Deadline(); ok {
		if err = s.SetDeadline(t); err != nil {
			return errors.Wrap(err, "set deadlines")
		}
	}

	var g errgroup.Group

	g.Go(func() error {
		ch := make(chan error, 1)

		go func() {
			b := new(bytes.Buffer)
			if _, err = io.Copy(b, io.LimitReader(s, 8)); err != nil {
				ch <- errors.Wrap(err, "read header")
				close(ch)
				return
			}

			conn.SetRemoteID(PeerID(binary.BigEndian.Uint64(b.Bytes())))
		}()

		var err error
		select {
		case err = <-ch:
		case <-c.Done():
			err = c.Err()
		}

		return errors.Wrap(err, "recv header")
	})

	g.Go(func() error {
		ch := make(chan error, 1)
		go func() {
			err := binary.Write(s, binary.BigEndian, local)
			ch <- errors.Wrap(err, "write")
			close(ch)
		}()

		var err error
		select {
		case err = <-ch:
		case <-c.Done():
			err = c.Err()
		}

		return errors.Wrap(err, "send header")
	})

	return g.Wait()
}

type bind struct {
	c context.Context
	Stream
}

func (b bind) Context() context.Context { return b.c }

// Bind context to a stream
func Bind(c context.Context, s Stream) Stream {
	return bind{c: c, Stream: s}
}
