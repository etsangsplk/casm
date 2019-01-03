package host

import (
	"context"

	"github.com/SentimensRG/ctx"
	casm "github.com/lthibault/casm/pkg"
	net "github.com/lthibault/casm/pkg/net"
	log "github.com/lthibault/log/pkg"
	"github.com/pkg/errors"
)

type basicHost struct {
	l log.Logger

	a net.Addr
	t *net.Transport

	*streamMux
	peers *peerStore
}

func (bh basicHost) log() log.Logger {
	return bh.l.WithFields(log.F{
		"id":         bh.a.ID(),
		"local_peer": bh.a,
	})
}

func (bh basicHost) Addr() net.Addr        { return bh.a }
func (bh basicHost) Network() Network      { return bh }
func (bh basicHost) Stream() StreamManager { return bh }

func (bh *basicHost) ListenAndServe(c context.Context, a net.Addr) error {
	bh.a = a // assign listen address

	c = log.Set(c, bh.log().WithLocus("listener"))

	l, err := bh.t.NewListener(a).Listen(c)
	if err != nil {
		return errors.Wrap(err, "listen")
	}
	ctx.Defer(c, func() { l.Close() })

	go bh.startAccepting(c, l)

	bh.log().Info("started host")
	return nil

}

func (bh basicHost) startAccepting(c context.Context, l *net.Listener) {
	var err error
	var conn *net.Conn

	for range ctx.Tick(c) {
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
		go bh.handle(c, conn.WithContext(log.Set(conn.Context(), l)))
	}
}

func (bh basicHost) handle(c context.Context, conn *net.Conn) {
	defer bh.Disconnect(conn.RemoteAddr())

	var err error
	var s *net.Stream
	for range ctx.Tick(ctx.Link(c, conn.Context())) {
		if s, err = conn.AcceptStream(); err != nil {
			bh.log().WithError(err).Warn("failed to accept stream")
			return
		}

		l := bh.log().WithFields(log.F{
			"remote_peer": s.RemoteAddr(),
			"stream":      s.StreamID(),
		})
		l.Debug("handling stream")

		go handleStream(bh, s.WithContext(log.Set(s.Context(), l)))
	}
}

func handleStream(h Handler, s *net.Stream) {
	var p path
	if err := p.RecvFrom(s); err != nil {
		log.Get(s.Context()).WithError(err).Debug("failed to read path")
	}

	h.Serve(stream{path: p.String(), Stream: s})
}

/*
	implment StreamManager
*/

func (bh basicHost) Open(a casm.Addresser, p string) (Stream, error) {
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

	if err = path(p).SendTo(s); err != nil {
		return nil, errors.Wrap(err, "write path")
	}

	l = l.WithField("path", p)
	l.Debug("header sent")

	return stream{
		path:   p,
		Stream: s.WithContext(log.Set(s.Context(), l)),
	}, nil
}

/*
	Implement Network
*/

func (bh basicHost) Connect(c context.Context, a casm.Addresser) error {
	if bh.a.ID() == a.Addr().ID() {
		return errors.New("cannot connect to self")
	}

	if _, connected := bh.peers.Get(a.Addr()); connected {
		return errors.Errorf("%s already connected", a.Addr().ID())
	}

	conn, err := bh.t.NewDialer(bh.a).Dial(c, a.Addr())
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
