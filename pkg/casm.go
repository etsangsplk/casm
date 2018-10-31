package casm

import (
	"context"

	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	"github.com/pkg/errors"
)

// Host is a logical machine in a compute cluster.  It acts both as a server and
// a client.  In the CASM expander-graph model, it is a vertex.
type Host interface {
	// ID() PeerID // TODO:  remove (redundant with addr)
	Addr() Addr
	Context() context.Context
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
func (h basicHost) Context() context.Context { return h.c }
func (h basicHost) Addr() Addr               { return h.a }
