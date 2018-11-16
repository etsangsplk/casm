package casm

import (
	"context"
	"encoding/binary"
	"io"
	"sync"

	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
)

func getID(c context.Context) PeerID   { return c.Value(keyID).(PeerID) }
func getHID(c context.Context) peer.ID { return c.Value(keyHID).(peer.ID) }

type yMap struct {
	sync.RWMutex
	m map[interface{}]context.Context
}

func newYMap() *yMap { return &yMap{m: make(map[interface{}]context.Context)} }

func (ym *yMap) Get(key interface{}) (c context.Context, ok bool) {
	ym.RLock()
	c, ok = ym.m[key]
	ym.RUnlock()
	return
}

func (ym *yMap) Put(c context.Context, id IDer, hid peer.ID) {
	c = context.WithValue(c, keyID, id.ID())
	c = context.WithValue(c, keyHID, hid)
	ym.put(c)
}

func (ym *yMap) put(c context.Context) {
	ym.Lock()
	ym.m[getID(c)] = c
	ym.m[getHID(c)] = c
	ym.Unlock()
}

func (ym *yMap) Del(key interface{}) {
	ym.Lock()
	c := ym.m[key]
	delete(ym.m, getID(c))
	delete(ym.m, getHID(c))
	ym.Unlock()
}

type contextHook struct {
	IDer
	c context.Context
	m *yMap
}

func newCtxHook(c context.Context, id IDer, ym *yMap) *contextHook {
	return &contextHook{c: c, IDer: id, m: ym}
}

func (h contextHook) Listen(net.Network, ma.Multiaddr)      {}
func (h contextHook) ListenClose(net.Network, ma.Multiaddr) {}
func (h contextHook) OpenedStream(net.Network, net.Stream)  {}
func (h contextHook) ClosedStream(net.Network, net.Stream)  {}

func writeID(w io.Writer, id IDer) (err error) {
	return binary.Write(w, binary.BigEndian, id.ID())
}

func readID(r io.Reader) (id PeerID, err error) {
	err = binary.Read(io.LimitReader(r, 8), binary.BigEndian, &id)
	return
}

func (h contextHook) Connected(_ net.Network, conn net.Conn) {
	s, err := conn.NewStream()
	if err != nil {
		conn.Close()
	}
	defer s.Reset()

	switch conn.Stat().Direction {
	case net.DirOutbound:
		if err = writeID(s, h); err != nil {
			conn.Close()
		}
	case net.DirInbound:
		if id, err := readID(s); err != nil {
			conn.Close()
		} else {
			h.m.Put(h.c, id, conn.RemotePeer())
		}
	case net.DirUnknown:
		panic("unknown direction")
	}
}

func (h contextHook) Disconnected(_ net.Network, conn net.Conn) {
	h.m.Del(HostLabel(conn.RemotePeer()))
}
