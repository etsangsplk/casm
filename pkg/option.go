package casm

import (
	"context"

	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	"github.com/pkg/errors"
)

// Host is a logical machine in a compute cluster.  It acts both as a server and
// a client.  In the CASM expander-graph model, it is a vertex.
type Host struct {
	h host.Host
	k uint8
}

// New Host whose lifetime is bound to the context c.
func New(c context.Context, opt ...Option) (h *Host, err error) {
	copt := defaultHostOpts()
	copt.Load(opt)

	popt := defaultP2pOpts()
	popt.Load(opt)

	h = new(Host)
	if h.h, err = libp2p.New(c, popt...); err != nil {
		err = errors.Wrap(err, "libp2p")
	}

	for _, o := range copt {
		if err = o.Apply(h); err != nil {
			break
		}
	}

	return
}
