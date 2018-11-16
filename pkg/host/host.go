package host

import (
	"context"
	"encoding/binary"
	"sync"
	"unsafe"

	"github.com/SentimensRG/ctx"
	"github.com/SentimensRG/ctx/mergectx"

	radix "github.com/armon/go-radix"
	casm "github.com/lthibault/casm/pkg"
	log "github.com/lthibault/casm/pkg/log"
	net "github.com/lthibault/casm/pkg/net"
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
	Open(context.Context, casm.Addresser, string) (net.Stream, error)
}

// Host is a logical machine in a compute cluster.  It acts both as a server and
// a client.  In the CASM expander-graph model, it is a vertex.
type Host interface {
	Context() context.Context
	Addr() net.Addr
	Network() Network
	Stream() StreamManager
	ListenAndServe(c context.Context) error
}

type basicHost struct {
	log   log.Logger
	c     context.Context
	id    casm.IDer
	addr  string
	mux   *mux
	peers *peerStore
	t     net.Transport
}

// New Host whose lifetime is bound to the context c.
func New(opt ...Option) Host {
	cfg := new(cfg)
	for _, fn := range defaultOpts(opt...) {
		fn(cfg)
	}

	return cfg.mkHost()
}

func (bh basicHost) Addr() net.Addr { return net.NewAddr(bh.id.ID(), bh.addr) }

func (bh basicHost) Network() Network {
	if bh.c == nil {
		panic(errors.New("host not started"))
	}
	return bh
}

func (bh basicHost) Stream() StreamManager {
	if bh.c == nil {
		panic(errors.New("host not started"))
	}
	return bh
}

func (bh basicHost) Context() context.Context {
	if bh.c == nil {
		panic(errors.New("host not started"))
	}
	return bh.c
}

func (bh *basicHost) ListenAndServe(c context.Context) error {
	bh.log = bh.log.WithFields(log.F{
		"id":         bh.id.ID(),
		"local_peer": bh.Addr(),
	})
	bh.log.Info("starting host")
	bh.c = c

	l, err := bh.t.Listen(bh.c, bh.Addr())
	if err != nil {
		return errors.Wrap(err, "listen")
	}
	ctx.Defer(bh.c, func() { l.Close() })

	go func() {
		for range ctx.Tick(bh.c) {
			if conn, err := l.Accept(bh.c); err != nil {
				bh.log.WithError(err).Warn("accept conn")
			} else if err = bh.peers.Add(conn); err != nil {
				bh.log.WithError(err).Warn("store peer")
			} else {
				ctx.Defer(conn.Context(), func() {
					bh.Disconnect(conn.Endpoint().Remote())
				})
			}
		}
	}()

	return nil
}

/*
	implment StreamManager
*/

func (bh basicHost) Register(path string, h net.Handler) {
	c := log.Set(bh.c, bh.log.WithLocus("mux"))
	bh.mux.Register(c, path, h)
}

func (bh basicHost) Unregister(path string) { bh.mux.Unregister(path) }

func (bh basicHost) Open(c context.Context, a casm.Addresser, path string) (s net.Stream, err error) {
	log := bh.log.WithFields(log.F{
		"remote_peer": a.Addr(),
		"path":        path,
	})

	conn, err := bh.peers.Get(a.Addr())
	if err != nil {
		return nil, errors.Wrap(err, "get peer")
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

func (bh basicHost) Connect(c context.Context, a casm.Addresser) (err error) {
	l := bh.log.WithField("remote_peer", a.Addr())
	l.Debug("connecting")
	defer l.IfNoErr(func(l log.Logger) {
		l.Debug("connected")
	}).Eval(err)

	c = log.Set(c, l.WithLocus("transport"))
	conn, err := bh.t.Dial(c, a.Addr())
	if err != nil {
		bh.log.WithField("addr", a.Addr()).WithError(err).Debug("connect")
		err = errors.Wrap(err, "transport")
	} else {
		err = errors.Wrap(bh.peers.Add(conn), "add peer")
	}

	return
}

func (bh basicHost) Disconnect(id casm.IDer) {
	if conn, ok := bh.peers.Del(id.ID()); ok {
		// TODO: log error
		conn.Close()
	}
}

type peerStore struct {
	sync.RWMutex
	m map[net.PeerID]net.Conn
}

func (p *peerStore) Add(conn net.Conn) error {
	p.Lock()
	defer p.Unlock()

	id := conn.Endpoint().Remote().ID()
	if _, ok := p.m[id]; ok {
		return errors.New("peer already connected")
	}
	p.m[id] = conn
	return nil
}

func (p *peerStore) Get(id casm.IDer) (c net.Conn, err error) {
	p.RLock()
	defer p.RUnlock()
	var ok bool
	if c, ok = p.m[id.ID()]; !ok {
		err = errors.New("not found")
	}
	return
}

func (p *peerStore) Del(id casm.IDer) (conn net.Conn, ok bool) {
	p.Lock()
	conn, ok = p.m[id.ID()]
	delete(p.m, id.ID())
	p.Unlock()
	return
}

type radixRouter radix.Tree

func (r *radixRouter) Insert(path string, b bind) {
	(*radix.Tree)(unsafe.Pointer(r)).Insert(path, b)
}

func (r *radixRouter) Delete(path string) (b bind, ok bool) {
	var v interface{}
	if v, ok = (*radix.Tree)(unsafe.Pointer(r)).Delete(path); ok {
		b = v.(bind)
	}
	return
}

func (r *radixRouter) Serve(s net.Stream) {
	h, ok := (*radix.Tree)(unsafe.Pointer(r)).Get(s.Path())
	if !ok {
		s.Close() // TODO:  implement Stream.CloseWithError ?
	}

	go h.(net.Handler).Serve(s)
}

type bind struct {
	c context.Context
	h net.Handler
}

func (b bind) Serve(s net.Stream) {
	c := mergectx.Link(b.c, s.Context())
	c = log.Set(c, log.Get(c).WithLocus("handler"))
	go b.h.Serve(net.Bind(c, s))
}

type mux struct {
	lock sync.RWMutex
	*radixRouter
}

func newMux() *mux {
	return &mux{radixRouter: (*radixRouter)(radix.New())}
}

func (m *mux) Register(c context.Context, path string, h net.Handler) {
	m.lock.Lock()
	log.Get(c).WithFields(log.F{"path": path, "handler": h}).Debug("registered")
	m.Insert(path, bind{c: c, h: h})
	m.lock.Unlock()
}

func (m *mux) Unregister(path string) {
	m.lock.Lock()
	if b, ok := m.Delete(path); ok {
		log.Get(b.c).Debug("unregistered")
	}
	m.lock.Unlock()
}

func (m *mux) Serve(s net.Stream) {
	m.lock.RLock()
	m.Serve(s)
	m.lock.RUnlock()
}
