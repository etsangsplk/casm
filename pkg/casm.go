// Package casm implements raw CASM hosts
package casm

import (
	"context"
	"sync"

	"github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"

	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
	"github.com/pkg/errors"
)

const (
	keyPID = "pid"
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
	sync.RWMutex
	m map[PeerID]peer.ID
}

func (m *idMap) Get(id IDer) (hid peer.ID, ok bool) {
	m.RLock()
	hid, ok = m.m[id.ID()]
	m.RUnlock()
	return
}

func (m *idMap) Set(id IDer, hid peer.ID) {
	m.Lock()
	m.m[id.ID()] = hid
	m.Unlock()
}

func (m *idMap) Del(id IDer) {
	m.Lock()
	delete(m.m, id.ID())
	m.Unlock()
}

type idMapHook struct {
	*idMap
	IDer
}

func newIDMapHook(idm *idMap, id IDer) idMapHook {
	return idMapHook{idMap: idm, IDer: id}
}

func (m idMapHook) Listen(net.Network, ma.Multiaddr)      {}
func (m idMapHook) ListenClose(net.Network, ma.Multiaddr) {}
func (m idMapHook) OpenedStream(net.Network, net.Stream)  {}
func (m idMapHook) ClosedStream(net.Network, net.Stream)  {}

func (m idMapHook) Connected(_ net.Network, conn net.Conn) {
	m.Set(m, conn.LocalPeer())
}

// called when a connection closed
func (m idMapHook) Disconnected(net net.Network, _ net.Conn) {
	defer net.StopNotify(m) // corresponds to AddNetHook in basicHost.Connect

	m.Lock()
	for k := range m.m {
		if k == m.ID() {
			delete(m.m, k)
		}
	}
	m.Unlock()
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

	h := &basicHost{c: c, idmap: &idMap{m: make(map[PeerID]peer.ID)}}
	if h.h, err = libp2p.New(c, popt...); err != nil {
		return nil, errors.Wrap(err, "libp2p")
	}

	pa := host.PeerInfoFromHost(h.h)
	h.a = &addr{IDer: NewID(), l: HostLabel(pa.ID), as: pa.Addrs}

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
		strm := newStream(bh.c, s)
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

	// pass host's context because context `c` is the stream-open context.  It
	// may contain timeouts.
	return newStream(bh.c, s), nil
}

/*
	Implement Network
*/

func (bh basicHost) Connect(c context.Context, a Addresser) error {
	bh.Hook().Add(newIDMapHook(bh.idmap, a.Addr()))

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
