package host

import (
	"context"
	"encoding/binary"
	"sync"
	"unsafe"

	"github.com/SentimensRG/ctx"

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
	l    log.Logger
	c    context.Context
	id   casm.IDer
	addr string
	*mux
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

func (bh basicHost) log() log.Logger { return bh.l }
func (bh basicHost) Addr() net.Addr  { return net.NewAddr(bh.id.ID(), bh.addr) }

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
	bh.c = c

	l, err := bh.t.Listen(c, bh.Addr())
	if err != nil {
		return errors.Wrap(err, "listen")
	}
	ctx.Defer(c, func() { l.Close() })

	go func() {
		for range ctx.Tick(c) {
			if conn, err := l.Accept(c); err != nil {
				bh.log().WithError(err).Warn("accept conn")
			} else if err = bh.peers.Add(conn); err != nil {
				bh.log().WithError(err).Warn("store peer")
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

func (bh basicHost) Open(c context.Context, a casm.Addresser, path string) (s net.Stream, err error) {
	conn, err := bh.peers.Get(a.Addr())
	if err != nil {
		return nil, errors.Wrap(err, "get peer")
	}

	cherr0 := make(chan error)
	cherr1 := make(chan error)

	go func() {
		var e error
		if s, e = conn.Stream().Open(); e != nil {
			bh.log().WithError(e).Warn("open stream")
			e = errors.Wrap(e, "open stream")
		}

		select {
		case <-c.Done():
		case cherr0 <- e:
		}
	}()

	go func() {
		select {
		case <-c.Done():
		case e := <-cherr0:
			if e == nil {
				e = binary.Write(s, binary.BigEndian, path)
			}

			select {
			case <-c.Done():
			case cherr1 <- errors.Wrap(e, "write path"):
				if e != nil {
					bh.log().WithError(e).Warn("write path")
					s.Close() // TODO:  CloseWithError
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
	conn, err := bh.t.Dial(c, a.Addr())
	if err != nil {
		bh.log().WithField("addr", a.Addr()).WithError(err).Debug("connect")
		return errors.Wrap(err, "transport")
	}

	return errors.Wrap(bh.peers.Add(conn), "add peer")
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

func (r *radixRouter) Insert(path string, h net.Handler) {
	(*radix.Tree)(unsafe.Pointer(r)).Insert(path, h)
}

func (r *radixRouter) Delete(path string) {
	(*radix.Tree)(unsafe.Pointer(r)).Delete(path)
}

func (r *radixRouter) ServeStream(s net.Stream) {
	h, ok := (*radix.Tree)(unsafe.Pointer(r)).Get(s.Path())
	if !ok {
		s.Close() // TODO:  implement Stream.CloseWithError ?
	}

	go h.(net.Handler).ServeStream(s)
}

type mux struct {
	lock sync.RWMutex
	*radixRouter
}

func (m *mux) Register(path string, h net.Handler) {
	m.lock.Lock()
	m.Insert(path, h)
	m.lock.Unlock()
}

func (m *mux) Unregister(path string) {
	m.lock.Lock()
	m.Delete(path)
	m.lock.Unlock()
}

func (m *mux) ServeStream(s net.Stream) {
	m.lock.RLock()
	m.ServeStream(s)
	m.lock.RUnlock()
}
