package casm

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	"github.com/pkg/errors"
)

func init() { rand.Seed(time.Now().UTC().UnixNano()) }

// ID is a unique identifier for a Node
type ID uint64

// NewID produces a random ID
func NewID() ID              { return ID(rand.Uint64()) }
func (id ID) String() string { return fmt.Sprintf("%016x", uint64(id)) }

// IDFromHex parses a hex string into a ID
func IDFromHex(x string) (id ID, err error) {
	var i uint64
	if i, err = strconv.ParseUint(x, 16, 64); err == nil {
		id = ID(i)
	}
	return
}

// ID satisfies the IDer interface
func (id ID) ID() ID { return id }

// Host is a logical machine in a compute cluster.  It acts both as a server and
// a client.  In the CASM expander-graph model, it is a vertex.
type Host struct {
	ID
	c context.Context
	h host.Host
}

// New Host whose lifetime is bound to the context c.
func New(c context.Context, opt ...Option) (h *Host, err error) {
	copt := defaultHostOpts()
	copt.Load(opt)

	popt := defaultP2pOpts()
	popt.Load(opt)

	h = &Host{c: c, ID: NewID()}
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

// Context to which the Host is bound
func (h Host) Context() context.Context { return h.c }
