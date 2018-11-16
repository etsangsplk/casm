package casm

import (
	"context"
	"encoding/binary"
	"sync"
	"unsafe"

	radix "github.com/armon/go-radix"
	net "github.com/lthibault/casm/pkg/net"
	"github.com/pkg/errors"
)

type basicHost struct {
	c    context.Context
	id   net.IDer
	addr string
	*mux
	peers *peerStore
	t     Transport
}

func mkHost() *basicHost {
	return &basicHost{
		mux:   &mux{radixRouter: (*radixRouter)(radix.New())},
		peers: &peerStore{m: make(map[net.PeerID]net.Conn)},
	}
}

// New Host whose lifetime is bound to the context c.
func New(c context.Context, opt ...Option) (Host, error) {
	var err error
	h := mkHost()
	for _, fn := range defaultOpts(opt...) {
		fn(h)
	}

	return h, err
}

// Addr on which to reach the Host
func (bh basicHost) Addr() net.Addr { return net.NewAddr(bh.id, bh.addr) }

// Context to which the Host is bound
func (bh basicHost) Context() context.Context { return bh.c }

func (bh basicHost) Network() Network      { return bh }
func (bh basicHost) Stream() StreamManager { return bh }

/*
	implment StreamManager
*/

func (bh basicHost) Open(c context.Context, a Addresser, path string) (s net.Stream, err error) {
	conn, err := bh.peers.Get(a.Addr())
	if err != nil {
		return nil, errors.Wrap(err, "get peer")
	}

	cherr0 := make(chan error)
	cherr1 := make(chan error)

	go func() {
		var e error
		if s, e = conn.Stream().Open(); e != nil {
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

func (bh basicHost) Connect(c context.Context, a Addresser) error {
	conn, err := dial(c, bh.t, a.Addr())
	if err != nil {
		return errors.Wrap(err, "dial")
	}

	return errors.Wrap(bh.peers.Add(conn), "add peer")
}

func (bh basicHost) Disconnect(id net.IDer) {
	bh.peers.Del(id.ID())
}

type peerStore struct {
	sync.RWMutex
	m map[PeerID]net.Conn
}

func (p *peerStore) Add(conn net.Conn) error {
	p.Lock()
	defer p.Unlock()

	id := conn.Endpoint().Remote().ID()
	if _, ok := p.m[id]; ok {
		return errors.New("already connected")
	}
	p.m[id] = conn
	return nil
}

func (p *peerStore) Get(id net.IDer) (c net.Conn, err error) {
	p.RLock()
	defer p.RUnlock()
	var ok bool
	if c, ok = p.m[id.ID()]; !ok {
		err = errors.New("not found")
	}
	return
}

func (p *peerStore) Del(id net.IDer) {
	p.Lock()
	delete(p.m, id.ID())
	p.Unlock()
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
