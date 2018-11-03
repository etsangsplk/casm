// Package casm implements raw CASM hosts
package casm

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"sync"

	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/hashbidimap"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p-peerstore"

	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
	"github.com/pkg/errors"
)

const (
	protoConnHdr = "/cxn"
)

// NetHookManager can add and remove nethooks from a Host
type NetHookManager interface {
	Add(NetHook)
	Remove(NetHook)
}

// Network manages raw connections
type Network interface {
	Connect(context.Context, Addresser) error
	Disconnect(IDer)
	Hook() NetHookManager
}

// StreamManager manages streams, which are multiplexed on top of raw connections
type StreamManager interface {
	Register(string, Handler)
	Unregister(string)
	Open(context.Context, Addresser, string) (Stream, error)
}

// Host is a logical machine in a compute cluster.  It acts both as a server and
// a client.  In the CASM expander-graph model, it is a vertex.
type Host interface {
	Context() context.Context
	Addr() Addr
	PeerAddr(IDer) (Addr, bool)
	Network() Network
	Stream() StreamManager
}

type idMap struct {
	IDer
	sync.RWMutex
	m maps.BidiMap
}

func newIDMap(id IDer) *idMap { return &idMap{IDer: id, m: hashbidimap.New()} }

func (m *idMap) Get(id IDer) (hid peer.ID, ok bool) {
	m.RLock()
	defer m.RUnlock()

	var v interface{}
	if v, ok = m.m.Get(id.ID()); ok {
		hid = v.(peer.ID)
	}

	return
}

func (m *idMap) GetKey(hid peer.ID) (id PeerID, ok bool) {
	m.RLock()
	defer m.RUnlock()

	var v interface{}
	if v, ok = m.m.GetKey(hid); ok {
		id = v.(PeerID)
	}

	return
}

func (m *idMap) Set(id IDer, hid peer.ID) {
	m.Lock()
	m.m.Put(id.ID(), hid)
	m.Unlock()
}

func (m *idMap) Del(id IDer) {
	m.Lock()
	m.m.Remove(id.ID())
	m.Unlock()
}

func (m *idMap) DelKey(v peer.ID) {
	m.Lock()
	if k, ok := m.m.GetKey(v); ok {
		m.m.Remove(k)
	}
	m.Unlock()
}

func (m *idMap) Listen(net.Network, ma.Multiaddr)      {}
func (m *idMap) ListenClose(net.Network, ma.Multiaddr) {}
func (m *idMap) OpenedStream(net.Network, net.Stream)  {}
func (m *idMap) ClosedStream(net.Network, net.Stream)  {}

func (m *idMap) Connected(_ net.Network, conn net.Conn) {
	s, err := conn.NewStream()
	if err != nil {
		conn.Close()
	}
	defer s.Reset()

	s.SetProtocol(protoConnHdr)

	switch conn.Stat().Direction {
	case net.DirOutbound:
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(m.ID()))
		if _, err = io.Copy(s, bytes.NewBuffer(b)); err != nil {
			conn.Close()
		}
	case net.DirInbound:
		r := io.LimitReader(s, 8)
		buf := new(bytes.Buffer)
		if _, err = io.Copy(buf, r); err != nil {
			conn.Close()
		}

		m.Set(PeerID(binary.BigEndian.Uint64(buf.Bytes())), conn.RemotePeer())
	case net.DirUnknown:
		panic("unknown direction")
	}
}

// called when a connection closed
func (m *idMap) Disconnected(_ net.Network, conn net.Conn) {
	m.DelKey(conn.RemotePeer())
}

type basicHost struct {
	a     Addr
	c     context.Context
	h     host.Host
	idmap *idMap
}

// New Host whose lifetime is bound to the context c.
func New(c context.Context, opt ...Option) (Host, error) {
	var err error

	copt := defaultHostOpts()
	copt.Load(opt)

	popt := defaultP2pOpts()
	popt.Load(opt)

	h := &basicHost{c: c, idmap: newIDMap(NewID())}

	if h.h, err = libp2p.New(c, popt...); err != nil {
		return nil, errors.Wrap(err, "libp2p")
	}

	pa := host.PeerInfoFromHost(h.h)
	h.a = &addr{IDer: h.idmap, l: HostLabel(pa.ID), as: pa.Addrs}
	h.Add(h.idmap)

	for _, o := range copt {
		if err = o.Apply(h); err != nil {
			break
		}
	}

	return h, err
}

// Context to which the Host is bound
func (bh basicHost) Context() context.Context { return bh.c }
func (bh basicHost) Addr() Addr               { return bh.a }
func (bh basicHost) PeerAddr(id IDer) (a Addr, ok bool) {
	var hid peer.ID
	if hid, ok = bh.idmap.Get(id); ok {
		a = &addr{IDer: id, l: HostLabel(hid), as: bh.h.Peerstore().Addrs(hid)}
	}
	return
}

func (bh basicHost) Network() Network      { return bh }
func (bh basicHost) Hook() NetHookManager  { return bh }
func (bh basicHost) Stream() StreamManager { return bh }

/*
	implment StreamManager
*/

func (bh basicHost) Register(path string, h Handler) {
	bh.h.SetStreamHandler(protocol.ID(path), func(s net.Stream) {
		id, ok := bh.idmap.GetKey(s.Conn().RemotePeer())
		if !ok {
			panic("should have PeerID in idmap")
		}

		strm := newStream(bh.c, id, s)
		defer strm.Close()
		h.ServeStream(strm)
	})
}

func (bh basicHost) Unregister(path string) {
	bh.h.RemoveStreamHandler(protocol.ID(path))
}

func (bh basicHost) Open(c context.Context, a Addresser, path string) (Stream, error) {
	s, err := bh.h.NewStream(c, peer.ID(a.Addr().Label()), protocol.ID(path))
	if err != nil {
		return nil, errors.Wrap(err, "libp2p")
	}

	id, ok := bh.idmap.GetKey(s.Conn().RemotePeer())
	if !ok {
		panic("should have PeerID in idmap")
	}

	// pass host's context because context `c` is the stream-open context.  It
	// may contain timeouts.
	return newStream(bh.c, id, s), nil
}

/*
	Implement Network
*/

func (bh basicHost) Connect(c context.Context, a Addresser) error {
	bh.idmap.Set(a.Addr(), peer.ID(a.Addr().Label()))

	return bh.h.Connect(c, peerstore.PeerInfo{
		ID:    peer.ID(a.Addr().Label()),
		Addrs: a.Addr().MultiAddrs(),
	})
}

func (bh basicHost) Disconnect(id IDer) {
	if a, ok := bh.PeerAddr(id); ok {
		bh.h.Network().ClosePeer(peer.ID(a.Addr().Label()))
		bh.h.Peerstore().ClearAddrs(peer.ID(a.Addr().Label())) // necessary?
	}
}

/*
	Implement NetHookManager
*/

func (bh basicHost) Add(h NetHook)    { bh.h.Network().Notify(h) }
func (bh basicHost) Remove(h NetHook) { bh.h.Network().StopNotify(h) }
