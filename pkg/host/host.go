package host

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"

	"github.com/SentimensRG/ctx"

	casm "github.com/lthibault/casm/pkg"
	net "github.com/lthibault/casm/pkg/net"
	log "github.com/lthibault/log/pkg"
	"github.com/pkg/errors"
)

// Network manages raw connections
type Network interface {
	Connect(context.Context, casm.Addresser) error
	Disconnect(casm.IDer)
}

// StreamManager manages streams, which are multiplexed on top of raw connections
type StreamManager interface {
	Register(string, net.Handler)
	Unregister(string)
	Open(casm.Addresser, string) (*net.Stream, error)
}

// Host is a logical machine in a compute cluster.  It acts both as a server and
// a client.  In the CASM expander-graph model, it is a vertex.
type Host interface {
	Context() context.Context
	Addr() net.Addr
	Network() Network
	Stream() StreamManager
	Start(c context.Context) error
}

type basicHost struct {
	log   log.Logger
	c     context.Context
	a     net.Addr
	mux   *streamMux
	peers *peerStore
	t     net.Transport
}

// New Host.  Pass options to override defaults.
func New(opt ...Option) Host {
	c := cfg{}
	for _, fn := range opt {
		fn(&c)
	}

	return mkHost(c)
}

func (bh basicHost) Addr() net.Addr { return bh.a }

func (bh basicHost) Network() Network {
	if bh.c == nil {
		panic(errors.New("host not started"))
	}
	return bh
}

func (bh basicHost) Stream() StreamManager { return bh }

func (bh basicHost) Context() context.Context {
	if bh.c == nil {
		panic(errors.New("host not started"))
	}
	return bh.c
}

func (bh *basicHost) Start(c context.Context) error {
	bh.log = bh.log.WithFields(log.F{
		"id":         bh.a.ID(),
		"local_peer": bh.a,
	})

	bh.c = log.Set(c, bh.log)
	c = log.Set(c, bh.log.WithLocus("transport"))

	l, err := bh.t.Listen(c, bh.a)
	if err != nil {
		return errors.Wrap(err, "listen")
	}

	ctx.Defer(bh.c, func() { l.Close() })
	go bh.startAccepting(l)

	bh.log.Info("started host")
	return nil
}

func (bh basicHost) startAccepting(l net.Listener) {
	var err error
	var conn *net.Conn

	for range ctx.Tick(bh.c) {
		if conn, err = l.Accept(); err != nil {
			bh.log.WithError(err).Warn("failed to accept conn")
			return
		}

		if !bh.peers.Add(conn) {
			bh.log.Error("peer already connected")
			conn.Close()
			return
		}

		l := bh.log.WithField("remote_peer", conn.RemoteAddr())

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
			bh.log.WithError(err).Warn("failed to accept stream")
			return
		}

		l := bh.log.WithFields(log.F{
			"remote_peer": s.RemoteAddr(),
			"stream":      s.StreamID(),
		})
		l.Debug("handling stream")

		c := log.Set(s.Context(), l)
		go bh.handleStream(s.WithContext(c))
	}
}

func (bh basicHost) handleStream(s *net.Stream) {
	var hdrLen uint16
	if err := binary.Read(s, binary.BigEndian, &hdrLen); err != nil {
		bh.log.WithError(err).Warn("failed to read stream header")
		return
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, io.LimitReader(s, int64(hdrLen))); err != nil {
		bh.log.WithError(err).Warn("failed to read stream path")
		return
	}

	c := log.Set(s.Context(), bh.log.WithFields(log.F{
		"locus": "handler",
		"path":  buf.String(),
	}))
	bh.mux.Serve(buf.String(), s.WithContext(c))
}

/*
	implment StreamManager
*/

func (bh basicHost) Register(path string, h net.Handler) {
	c := log.Set(bh.c, bh.log.WithLocus("mux"))
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

	l := bh.log.WithField("stream", s.StreamID())
	l.Debug("stream opened")

	var hdr = uint16(len(path))
	if err = binary.Write(s, binary.BigEndian, hdr); err != nil {
		return nil, errors.Wrap(err, "write header")
	}

	if err = binary.Write(s, binary.BigEndian, []byte(path)); err != nil {
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

	conn, err := bh.t.Dial(c, a.Addr())
	if err != nil {
		return errors.Wrap(err, "transport")
	}

	if !bh.peers.Add(conn) {
		conn.Close()
		return errors.New("peer already connected")
	}

	bh.log.WithField("remote_peer", conn.RemoteAddr()).Debug("connected to peer")
	return nil
}

func (bh basicHost) Disconnect(id casm.IDer) {
	if conn, ok := bh.peers.Del(id.ID()); ok {
		conn.Close()
	}
}
