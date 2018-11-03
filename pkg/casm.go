// Package casm implements raw CASM hosts
package casm

import (
	"context"

	"github.com/libp2p/go-libp2p-peerstore"

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
	PeerAddr(HostLabel) (Addr, bool)
	Context() context.Context
	RegisterStreamHandler(string, Handler)
	UnregisterStreamHandler(string)
	OpenStream(context.Context, Addresser, string) (Stream, error)
	Connect(context.Context, Addresser) error
	Disconnect(Addresser)
}

type basicHost struct {
	a Addr
	c context.Context
	h host.Host
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
func (bh basicHost) PeerAddr(l HostLabel) (Addr, bool) {
	v, err := bh.h.Peerstore().Get(peer.ID(l), keyPID)
	if err != nil {
		return nil, false
	}

	pi := bh.h.Peerstore().PeerInfo(peer.ID(l))
	return &addr{IDer: v.(PeerID), l: l, as: pi.Addrs}, true
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
	}); err != nil {
		return
	}

	// This _should_ get cleared on call to ClearAddrs or peer disconnection...
	return bh.h.Peerstore().Put(peer.ID(a.Addr().Label()), keyPID, a.Addr().ID())
}

// Disconnect from a peer
func (bh basicHost) Disconnect(a Addresser) {
	bh.h.Peerstore().ClearAddrs(peer.ID(a.Addr().Label()))
}
