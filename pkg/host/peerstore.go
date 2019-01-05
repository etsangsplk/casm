package host

import (
	"context"
	"sync"

	casm "github.com/lthibault/casm/pkg"
	net "github.com/lthibault/casm/pkg/net"
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

type cxnTable map[net.PeerID]cxn

func (t cxnTable) Add(conn cxn) (added bool) {
	id := conn.RemoteAddr().ID()
	if _, ok := t.Get(id); ok {
		return false
	}
	t[id] = conn
	return true
}

func (t cxnTable) Get(id net.PeerID) (c cxn, found bool) {
	c, found = t[id]
	return
}

func (t cxnTable) Del(id net.PeerID) (conn cxn, found bool) {
	conn, found = t[id]
	delete(t, id)
	return
}

type peerStore struct {
	sync.RWMutex
	t cxnTable
}

func newPeerStore() *peerStore { return new(peerStore).Reset() }

func (p *peerStore) Retrieve(id casm.IDer) (conn cxn, found bool) {
	p.RLock()
	conn, found = p.t.Get(id.ID())
	p.RUnlock()

	return
}

func (p *peerStore) StoreOrClose(conn cxn) (stored bool) {

	p.Lock()
	if stored = p.t.Add(conn); !stored {
		conn.Close()
	}
	p.Unlock()

	return
}

func (p *peerStore) DropAndClose(id casm.IDer) {
	p.Lock()
	if conn, ok := p.t.Del(id.ID()); ok {
		conn.Close()
	}
	p.Unlock()
}

func (p *peerStore) Contains(id casm.IDer) (found bool) {
	p.RLock()
	_, found = p.t.Get(id.ID())
	p.RUnlock()
	return
}

func (p *peerStore) Reset() *peerStore {
	p.Lock()
	p.t = make(map[net.PeerID]cxn)
	p.Unlock()
	return p
}
