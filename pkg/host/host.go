package host

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"sync"

	"github.com/SentimensRG/ctx"
	radix "github.com/armon/go-radix"

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
	Open(context.Context, casm.Addresser, string) (*net.Stream, error)
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
	mux   *mux
	peers *peerStore
	t     *net.Transport
}

// New Host whose lifetime is bound to the context c.
func New(opt ...Option) Host {
	cfg := new(cfg)
	for _, fn := range opt {
		fn(cfg)
	}

	return cfg.mkHost()
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
	bh.log.Info("starting host")
	bh.c = log.Set(c, bh.log)
	c = log.Set(c, bh.log.WithLocus("transport"))

	l, err := bh.t.Listen(c, bh.a)
	if err != nil {
		return errors.Wrap(err, "listen")
	}

	ctx.Defer(bh.c, func() { l.Close() })
	go bh.startAccepting(l)

	return nil
}

func (bh basicHost) startAccepting(l net.Listener) {
	listenLog := bh.log.WithLocus("listener")
	listenCtx := log.Set(bh.c, listenLog)

	var err error
	var conn *net.Conn

	bh.log.Debug("listening")
	for range ctx.Tick(bh.c) {
		if conn, err = l.Accept(listenCtx); err != nil {
			bh.log.WithError(err).Warn("accept conn")
			return
		}

		if !bh.peers.Add(conn) {
			bh.log.Error("peer already connected")
			conn.Close()
			return
		}

		l := bh.log.WithField("remote_peer", conn.Endpoint().Remote())
		l.Debug("handling connection")

		go bh.handle(conn.WithContext(log.Set(conn.Context(), l)))
	}
}

func (bh basicHost) handle(conn *net.Conn) {
	defer bh.Disconnect(conn.Endpoint().Remote())

	var err error
	var s *net.Stream
	for range ctx.Tick(ctx.Link(bh.c, conn.Context())) {
		if s, err = conn.Stream().Accept(); err != nil {
			bh.log.WithError(err).Warn("accept stream")
			return
		}

		bh.log.WithField("remote_peer", conn.Endpoint().Remote()).Debug("connected")
		go bh.handleStream(s)
	}
}

func (bh basicHost) handleStream(s *net.Stream) {
	var hdrLen uint16
	if err := binary.Read(s, binary.BigEndian, &hdrLen); err != nil {
		bh.log.WithError(err).Warn("read stream header")
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, io.LimitReader(s, int64(hdrLen))); err != nil {
		bh.log.WithError(err).Warn("read stream path")
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

func (bh basicHost) Open(c context.Context, a casm.Addresser, path string) (s *net.Stream, err error) {
	log := bh.log.WithFields(log.F{
		"remote_peer": a.Addr(),
		"path":        path,
	})

	conn, ok := bh.peers.Get(a.Addr())
	if !ok {
		return nil, errors.New("peer not connected")
	}

	cherr0 := make(chan error)
	cherr1 := make(chan error)

	go func() {
		ch := make(chan error, 1)
		go func() {
			var e error
			if s, e = conn.Stream().Open(); e != nil {
				e = errors.Wrap(e, "open stream")
			}
			ch <- e
		}()

		select {
		case <-c.Done():
		case e := <-ch:
			select {
			case <-c.Done():
			case cherr0 <- e:
			}
		}
	}()

	go func() {
		ch := make(chan error, 1)

		select {
		case <-c.Done():
		case e := <-cherr0:
			if e == nil {
				go func() {
					if e = binary.Write(s, binary.BigEndian, path); e != nil {
						log.WithError(e).Warn("write path")
						e = errors.Wrap(e, "write path")
						s.Close() // TODO:  CloseWithError
					}
					ch <- e
				}()
			}

			select {
			case <-c.Done():
			case e := <-ch:
				select {
				case <-c.Done():
				case cherr1 <- e:
				}
			}
		}
	}()

	select {
	case <-c.Done():
		err = c.Err()
	case err = <-cherr1:
	}

	return
}

/*
	Implement Network
*/

func (bh basicHost) Connect(c context.Context, a casm.Addresser) error {

	if _, connected := bh.peers.Get(a.Addr()); connected {
		return errors.Errorf("%s already connected", a.Addr().ID())
	}

	conn, err := bh.t.Dial(c, bh.a.ID(), a.Addr())
	if err != nil {
		return errors.Wrap(err, "transport")
	}

	if !bh.peers.Add(conn) {
		conn.Close()
		err = errors.New("peer already connected")
	}

	bh.log.WithField("remote_peer", conn.Endpoint().Remote()).Debug("connected")
	return nil
}

func (bh basicHost) Disconnect(id casm.IDer) {
	if conn, ok := bh.peers.Del(id.ID()); ok {
		// TODO: log error
		conn.Close()
	}
}

type peerStore struct {
	sync.RWMutex
	m map[net.PeerID]*net.Conn
}

func (p *peerStore) Add(conn *net.Conn) bool {
	p.Lock()
	defer p.Unlock()

	id := conn.Endpoint().Remote().ID()
	if _, ok := p.m[id]; ok {
		return false
	}
	p.m[id] = conn
	return true
}

func (p *peerStore) Get(id casm.IDer) (c *net.Conn, ok bool) {
	p.RLock()
	c, ok = p.m[id.ID()]
	p.RUnlock()
	return
}

func (p *peerStore) Del(id casm.IDer) (conn *net.Conn, ok bool) {
	p.Lock()
	conn, ok = p.m[id.ID()]
	delete(p.m, id.ID())
	p.Unlock()
	return
}

type mux struct {
	lock sync.RWMutex
	log  log.Logger
	r    *radix.Tree
}

func newMux(l log.Logger) *mux { return &mux{log: l, r: radix.New()} }

func (m *mux) Register(c context.Context, path string, h net.Handler) {
	m.lock.Lock()
	m.log.WithField("path", path).Debug("registered")
	m.r.Insert(path, h)
	m.lock.Unlock()
}

func (m *mux) Unregister(path string) {
	m.lock.Lock()
	if _, ok := m.r.Delete(path); ok {
		m.log.WithField("path", path).Debug("unregistered")
	}
	m.lock.Unlock()
}

func (m *mux) Serve(path string, s *net.Stream) {
	m.lock.RLock()
	if v, ok := m.r.Get(path); ok {
		go v.(net.Handler).Serve(s)
	}
	m.lock.RUnlock()
}
