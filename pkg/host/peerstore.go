package host

import (
	"sync"

	casm "github.com/lthibault/casm/pkg"
	net "github.com/lthibault/casm/pkg/net"
)

type peerStore struct {
	sync.RWMutex
	m map[net.PeerID]*net.Conn
}

func newPeerStore() *peerStore {
	return &peerStore{m: make(map[net.PeerID]*net.Conn)}
}

func (p *peerStore) Add(conn *net.Conn) bool {
	p.Lock()
	defer p.Unlock()

	id := conn.RemoteAddr().ID()
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
