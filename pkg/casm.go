// Package casm implements raw CASM hosts
package casm

import (
	"context"
	"log"
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

// Host is a logical machine in a compute cluster.  It acts both as a server and
// a client.  In the CASM expander-graph model, it is a vertex.
type Host interface {
	Addr() Addr
	PeerAddr(IDer) (Addr, bool)
	Context() context.Context
	RegisterStreamHandler(string, Handler)
	UnregisterStreamHandler(string)
	OpenStream(context.Context, Addresser, string) (Stream, error)
	Connect(context.Context, Addresser) error
	Disconnect(Addresser)
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

func (m *idMap) Listen(net.Network, ma.Multiaddr)      {}
func (m *idMap) ListenClose(net.Network, ma.Multiaddr) {}
func (m *idMap) OpenedStream(net.Network, net.Stream)  {}
func (m *idMap) ClosedStream(net.Network, net.Stream)  {}
func (m *idMap) Connected(_ net.Network, conn net.Conn) {
	log.Println(conn.Stat().Extra)
}

// called when a connection closed
func (m *idMap) Disconnected(_ net.Network, conn net.Conn) {
	m.Lock()
	defer m.Unlock()

	hid := conn.RemotePeer()
	for k, v := range m.m {
		if v == hid {
			delete(m.m, k)
		}
	}
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

	h := &basicHost{c: c}
	if h.h, err = libp2p.New(c, popt...); err != nil {
		return nil, errors.Wrap(err, "libp2p")
	}

	idmap := &idMap{m: make(map[PeerID]peer.ID)}
	h.idmap = idmap
	h.h.Network().Notify(h.idmap)

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

// RegisterStreamHandler
func (bh basicHost) RegisterStreamHandler(path string, h Handler) {
	bh.h.SetStreamHandler(protocol.ID(path), func(s net.Stream) {
		strm := newStream(bh.c, s)
		defer strm.Close()
		h.ServeStream(strm)
	})
}

// UnregisterStreamHandler
func (bh basicHost) UnregisterStreamHandler(path string) {
	bh.h.RemoveStreamHandler(protocol.ID(path))
}

// OpenStream
func (bh basicHost) OpenStream(c context.Context, a Addresser, path string) (Stream, error) {
	s, err := bh.h.NewStream(c, peer.ID(a.Addr().Label()), protocol.ID(path))
	if err != nil {
		return nil, errors.Wrap(err, "libp2p")
	}

	// pass host's context because context `c` is the stream-open context.  It
	// may contain timeouts.
	return newStream(bh.c, s), nil
}

// Connect to a peer
func (bh basicHost) Connect(c context.Context, a Addresser) (err error) {
	if err = bh.h.Connect(c, peerstore.PeerInfo{
		ID:    peer.ID(a.Addr().Label()),
		Addrs: a.Addr().MultiAddrs(),
	}); err == nil {
		bh.idmap.Set(a.Addr(), peer.ID(a.Addr().Label()))
	}

	return
}

// Disconnect from a peer
func (bh basicHost) Disconnect(a Addresser) {
	bh.h.Network().ClosePeer(peer.ID(a.Addr().Label()))
	bh.h.Peerstore().ClearAddrs(peer.ID(a.Addr().Label())) // necessary?
}
