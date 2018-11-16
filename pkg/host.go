package casm

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

type basicHost struct {
	c context.Context
	*addr
	peers *peerStore
	t     Transport
}

func mkHost() *basicHost {
	return &basicHost{
		peers: &peerStore{m: make(map[PeerID]Conn)},
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

// Context to which the Host is bound
func (bh basicHost) Context() context.Context { return bh.c }

func (bh basicHost) Network() Network      { return bh }
func (bh basicHost) Stream() StreamManager { return bh }

/*
	implment StreamManager
*/

func (bh basicHost) Register(path string, h Handler) {
	panic("Register NOT IMPLEMENTED")
}

func (bh basicHost) Unregister(path string) {
	panic("Unregister NOT IMPLEMENTED")
}

func (bh basicHost) Open(c context.Context, a Addresser, path string) (Stream, error) {
	panic("Open NOT IMPLEMENTED")
}

/*
	Implement Network
*/

func (bh basicHost) Connect(c context.Context, a Addresser) error {
	conn, err := Dial(c, bh.t, a.Addr())
	if err != nil {
		return errors.Wrap(err, "dial")
	}

	return errors.Wrap(bh.peers.Add(conn), "add peer")
}

func (bh basicHost) Disconnect(id IDer) {
	bh.peers.Del(id.ID())
}

type peerStore struct {
	sync.RWMutex
	m map[PeerID]Conn
}

func (p *peerStore) Add(conn Conn) error {
	p.Lock()
	defer p.Unlock()

	id := conn.Endpoint().Remote().ID()
	if _, ok := p.m[id]; ok {
		return errors.New("already connected")
	}
	p.m[id] = conn
	return nil
}

func (p *peerStore) Get(id IDer) (c Conn, err error) {
	p.RLock()
	defer p.RUnlock()
	var ok bool
	if c, ok = p.m[id.ID()]; !ok {
		err = errors.New("not found")
	}
	return
}

func (p *peerStore) Del(id IDer) {
	p.Lock()
	delete(p.m, id.ID())
	p.Unlock()
}
