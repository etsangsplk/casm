// Package host implements CASM's p2p host model
package host

import (
	"context"
	"io"

	"github.com/SentimensRG/ctx"
	casm "github.com/lthibault/casm/pkg"
	net "github.com/lthibault/casm/pkg/net"
	log "github.com/lthibault/log/pkg"
	"github.com/pkg/errors"
)

var (
	// ErrAlreadyConnected indicates that a connection attempt failed because
	// a connection to the remote Host already exists.
	ErrAlreadyConnected = errors.New("already connected")
)

// Host is a logical machine in a compute cluster.  It acts both as a server and
// a client.
type Host struct {
	l log.Logger

	a net.Addr
	t *net.Transport

	*streamMux
	peers *peerStore
}

// New Host.  Pass options to override defaults.
func New(opt ...Option) *Host {
	h := new(Host)

	for _, fn := range setDefaultOpts(opt) {
		fn(h)
	}

	h.streamMux = newStreamMux(h.l.WithLocus("mux"))
	h.peers = newPeerStore()
	return h
}

func (h Host) log() log.Logger {
	return h.l.WithFields(log.F{
		"id":         h.a.ID(),
		"local_peer": h.a,
	})
}

// Addr where the host can be reached
func (h Host) Addr() net.Addr { return h.a }

// Start the Host
func (h *Host) Start(c context.Context, a net.Addr) error {
	h.a = a // assign listen address

	c = log.Set(c, h.log().WithLocus("listener"))

	l, err := h.t.NewListener(a).Listen(c)
	if err != nil {
		return errors.Wrap(err, "listen")
	}
	ctx.Defer(c, h.halter(l))

	go h.startAccepting(c, l)

	h.log().Info("started host")
	return nil

}

func (h Host) halter(c io.Closer) func() {
	return func() {
		if err := c.Close(); err != nil {
			h.log().WithError(err).Fatal("unclean shutdown")
		} else {
			h.log().Warn("halted")
		}

		h.a = nil
		h.peers.Reset()
	}
}

func (h Host) startAccepting(c context.Context, l *net.Listener) {
	var err error
	var conn *net.Conn

	for range ctx.Tick(c) {
		if conn, err = l.Accept(); err != nil {
			select {
			case <-c.Done():
			default:
				h.log().WithError(err).Warn("failed to accept conn")
			}

			return
		}

		if !h.peers.Store(conn) {
			h.log().WithField("error", "already connected").Debug("closed connection")
			return
		}

		go h.handle(c, h.bindConnLogger(conn))
	}
}

func (h Host) bindConnLogger(conn *net.Conn) *net.Conn {
	return conn.WithContext(log.Set(
		conn.Context(),
		h.log().WithField("remote_peer", conn.RemoteAddr()),
	))
}

func (h Host) handle(c context.Context, conn *net.Conn) {
	log.Get(conn.Context()).Debug("connected")
	defer h.Disconnect(conn.RemoteAddr())

	var err error
	var s *net.Stream
	for range ctx.Tick(ctx.Link(c, conn.Context())) {
		if s, err = conn.AcceptStream(); err != nil {
			select {
			case <-c.Done():
			case <-conn.Context().Done():
			default:
				h.log().WithError(err).Error("failed to accept stream")
			}
			return
		}

		go handleStream(h, h.bindStreamLogger(s))
	}
}

func (h Host) bindStreamLogger(s *net.Stream) *net.Stream {
	return s.WithContext(log.Set(
		s.Context(),
		h.log().WithFields(log.F{
			"remote_peer": s.RemoteAddr(),
			"stream":      s.StreamID()},
		)),
	)
}

func handleStream(h Handler, s *net.Stream) {
	log.Get(s.Context()).Debug("stream accepted")

	var p streamPath
	if err := p.RecvFrom(s); err != nil {
		log.Get(s.Context()).WithError(err).Debug("failed to read path")
	}

	h.Serve(stream{path: p.String(), Stream: s})
}

// Open a stream. The peer must already be connected.
func (h Host) Open(a casm.Addresser, path string) (Stream, error) {
	conn, ok := h.peers.Retrieve(a.Addr())
	if !ok {
		return nil, errors.New("peer not found")
	}

	s, err := conn.OpenStream()
	if err != nil {
		return nil, errors.Wrap(err, "open stream")
	}

	if err = streamPath(path).SendTo(s); err != nil {
		return nil, errors.Wrap(err, "write path")
	}

	return h.bindStream(s, path), nil
}

func (h Host) bindStream(s *net.Stream, path string) stream {
	return stream{
		path: path,
		Stream: s.WithContext(log.Set(
			s.Context(),
			h.log().WithFields(log.F{"stream": s.StreamID(), "path": path}),
		)),
	}
}

/*
	Implement Network
*/

// Connect to a remote host.
func (h Host) Connect(c context.Context, a casm.Addresser) error {
	switch {
	case h.a == nil:
		return errors.New("host not started")
	case h.a.ID() == a.Addr().ID():
		return errors.New("cannot connect to self")
	case h.peers.Contains(a.Addr()):
		return ErrAlreadyConnected
	default:
		conn, err := h.dialAndStore(c, a.Addr())
		if err != nil {
			return err
		}

		go h.handle(c, conn)

		return nil
	}
}

func (h Host) dialAndStore(c context.Context, a net.Addr) (*net.Conn, error) {
	conn, err := h.t.NewDialer(h.a).Dial(c, a.Addr())
	if err != nil {
		return nil, errors.Wrap(err, "dial")
	}

	if !h.peers.Store(conn) {
		return nil, ErrAlreadyConnected
	}

	return h.bindConnLogger(conn), nil
}

// Disconnect from a remote host.
func (h Host) Disconnect(id casm.IDer) { h.peers.Drop(id) }
