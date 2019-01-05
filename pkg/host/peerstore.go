package host

import (
	"context"
	"sync"

	casm "github.com/lthibault/casm/pkg"
	net "github.com/lthibault/casm/pkg/net"
	"github.com/pkg/errors"
)

// cxn is a logical connection to a remote peer.
type cxn interface {
	Context() context.Context
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	AcceptStream() (*net.Stream, error)
	OpenStream() (*net.Stream, error)
	WithContext(context.Context) *net.Conn
}

type peerMap map[net.PeerID]cxn

func (m peerMap) Add(conn cxn) bool {
	id := conn.RemoteAddr().ID()
	if _, ok := m.Get(id); ok {
		return false
	}
	m[id] = conn
	return true
}

func (m peerMap) Get(id net.PeerID) (c cxn, ok bool) {
	c, ok = m[id]
	return
}

func (m peerMap) Del(id net.PeerID) (conn cxn, ok bool) {
	conn, ok = m[id]
	delete(m, id)
	return
}

type peerStore struct {
	sync.RWMutex
	m peerMap
}

func newPeerStore() *peerStore { return new(peerStore).Reset() }

func (p *peerStore) Retrieve(id casm.IDer) (conn cxn, err error) {
	var ok bool

	p.RLock()
	if conn, ok = p.m.Get(id.ID()); !ok {
		err = errors.New("peer not found")
	}
	p.RUnlock()

	return
}

func (p *peerStore) Store(conn cxn) (err error) {
	var drop bool

	p.Lock()

	if drop = p.m.Add(conn); drop {
		conn.Close()
		err = errors.New("peer already connected")
	}

	p.Unlock()
	return
}

func (p *peerStore) Drop(id casm.IDer) {
	p.Lock()
	if conn, ok := p.m.Del(id.ID()); ok {
		conn.Close()
	}
	p.Unlock()
}

func (p *peerStore) Contains(id casm.IDer) (contained bool) {
	p.RLock()
	_, contained = p.m.Get(id.ID())
	p.RUnlock()
	return
}

func (p *peerStore) Reset() *peerStore {
	p.Lock()
	p.m = make(map[net.PeerID]cxn)
	p.Unlock()
	return p
}
