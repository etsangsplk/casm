package host

import (
	"context"
	"sync"

	casm "github.com/lthibault/casm/pkg"
	net "github.com/lthibault/casm/pkg/net"
)

// cxn is a logical connection to a remote peer.
type cxn interface {
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	AcceptStream() (*net.Stream, error)
	OpenStream() (*net.Stream, error)
	WithContext(context.Context) *net.Conn
}

type peerStore struct {
	sync.RWMutex
	m map[net.PeerID]cxn
}

func newPeerStore() *peerStore {
	return &peerStore{m: make(map[net.PeerID]cxn)}
}

func (p *peerStore) Add(conn cxn) bool {
	p.Lock()
	defer p.Unlock()

	id := conn.RemoteAddr().ID()
	if _, ok := p.m[id]; ok {
		return false
	}
	p.m[id] = conn
	return true
}

func (p *peerStore) Get(id casm.IDer) (c cxn, ok bool) {
	p.RLock()
	c, ok = p.m[id.ID()]
	p.RUnlock()
	return
}

func (p *peerStore) Del(id casm.IDer) (conn cxn, ok bool) {
	p.Lock()
	conn, ok = p.m[id.ID()]
	delete(p.m, id.ID())
	p.Unlock()
	return
}
