package host

import (
	"context"

	"github.com/SentimensRG/ctx"
	casm "github.com/lthibault/casm/pkg"
	net "github.com/lthibault/casm/pkg/net"
	log "github.com/lthibault/log/pkg"
	"github.com/pkg/errors"
)

type ctxKey uint8

const (
	keyAddr ctxKey = iota
	keyLog
)

type basicHost struct {
	c context.Context
	t net.TransportFactory

	mux   *streamMux
	peers *peerStore
}

func (bh basicHost) Addr() net.Addr {
	return bh.c.Value(keyAddr).(net.Addr)
}

func (bh basicHost) Network() Network {
	if bh.c == nil {
		panic(errors.New("host not started"))
	}
	return bh
}

func (bh basicHost) log() log.Logger { return bh.c.Value(keyLog).(log.Logger) }

func (bh basicHost) setLog(l log.Logger) { log.Set(bh.c, l) }

func (bh basicHost) Stream() StreamManager { return bh }

func (bh basicHost) Context() context.Context {
	if bh.c == nil {
		panic(errors.New("host not started"))
	}
	return bh.c
}

func (bh *basicHost) ListenAndServe(c context.Context, a net.Addr) error {
	c = context.WithValue(c, keyAddr, a)
	bh.setLog(bh.log().WithFields(log.F{
		"id":         a.ID(),
		"local_peer": a,
	}))

	t, err := bh.t.NewTransport(a)
	if err != nil {
		return errors.Wrap(err, "init transport")
	}

	return bh.listenAndServe(t)
}

func (bh basicHost) listenAndServe(t net.Transport) error {
	c := log.Set(bh.c, bh.log().WithLocus("transport"))

	l, err := t.Listen(c)
	if err != nil {
		return errors.Wrap(err, "listen")
	}
	ctx.Defer(bh.c, func() { l.Close() })

	go bh.startAccepting(l)

	bh.log().Info("started host")
	return nil
}

func (bh basicHost) startAccepting(l net.Listener) {
	var err error
	var conn *net.Conn

	for range ctx.Tick(bh.c) {
		if conn, err = l.Accept(); err != nil {
			bh.log().WithError(err).Warn("failed to accept conn")
			return
		}

		if !bh.peers.Add(conn) {
			bh.log().Error("peer already connected")
			conn.Close()
			return
		}

		l := bh.log().WithField("remote_peer", conn.RemoteAddr())

		l.Debug("connection accepted")
		go bh.handle(conn.WithContext(log.Set(conn.Context(), l)))
	}
}

func (bh basicHost) handle(conn *net.Conn) {
	defer bh.Disconnect(conn.RemoteAddr())

	var err error
	var s *net.Stream
	for range ctx.Tick(ctx.Link(bh.c, conn.Context())) {
		if s, err = conn.AcceptStream(); err != nil {
			bh.log().WithError(err).Warn("failed to accept stream")
			return
		}

		l := bh.log().WithFields(log.F{
			"remote_peer": s.RemoteAddr(),
			"stream":      s.StreamID(),
		})
		l.Debug("handling stream")

		c := log.Set(s.Context(), l)
		go bh.handleStream(s.WithContext(c))
	}
}

func (bh basicHost) handleStream(s *net.Stream) {
	var p net.Path
	if _, err := p.ReadFrom(s); err != nil {
		bh.log().WithError(err).Error("failed to read stream path")
		return
	}

	l := bh.log().WithFields(log.F{"locus": "handler", "path": p})
	c := log.Set(s.Context(), l)

	bh.mux.Serve(p.String(), s.WithContext(c))
}

/*
	implment StreamManager
*/

func (bh basicHost) Register(path string, h net.Handler) {
	c := log.Set(bh.c, bh.log().WithLocus("mux"))
	bh.mux.Register(c, path, h)
}

func (bh basicHost) Unregister(path string) { bh.mux.Unregister(path) }

func (bh basicHost) Open(a casm.Addresser, path string) (*net.Stream, error) {
	conn, ok := bh.peers.Get(a.Addr())
	if !ok {
		return nil, errors.New("peer not connected")
	}

	s, err := conn.OpenStream()
	if err != nil {
		return nil, errors.Wrap(err, "open stream")
	}

	l := bh.log().WithField("stream", s.StreamID())
	l.Debug("stream opened")

	if _, err = net.Path(path).WriteTo(s); err != nil {
		return nil, errors.Wrap(err, "write path")
	}

	l = l.WithField("path", path)
	l.Debug("header sent")

	c := log.Set(s.Context(), l)
	return s.WithContext(c), nil
}

/*
	Implement Network
*/

func (bh basicHost) Connect(c context.Context, a casm.Addresser) error {
	if _, connected := bh.peers.Get(a.Addr()); connected {
		return errors.Errorf("%s already connected", a.Addr().ID())
	}

	t, err := bh.t.NewTransport(c.Value(keyAddr).(net.Addr))
	if err != nil {
		return errors.Wrap(err, "init transport")
	}

	conn, err := t.Dial(c, a.Addr())
	if err != nil {
		return errors.Wrap(err, "dial")
	}

	if !bh.peers.Add(conn) {
		conn.Close()
		return errors.New("peer already connected")
	}

	bh.log().WithField("remote_peer", conn.RemoteAddr()).Debug("connected to peer")
	return nil
}

func (bh basicHost) Disconnect(id casm.IDer) {
	if conn, ok := bh.peers.Del(id.ID()); ok {
		conn.Close()
	}
}
