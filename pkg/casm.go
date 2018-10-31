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

type IDer interface {
	ID() PeerID
}

// PeerID is a unique identifier for a Node
type PeerID uint64

// NewID produces a random PeerID
func NewID() PeerID              { return PeerID(rand.Uint64()) }
func (id PeerID) String() string { return fmt.Sprintf("%016x", uint64(id)) }

// IDFromHex parses a hex string into a PeerID
func IDFromHex(x string) (id PeerID, err error) {
	var i uint64
	if i, err = strconv.ParseUint(x, 16, 64); err == nil {
		id = PeerID(i)
	}
	return
}

// ID satisfies the IDer interface
func (id PeerID) ID() PeerID { return id }

// Host is a logical machine in a compute cluster.  It acts both as a server and
// a client.  In the CASM expander-graph model, it is a vertex.
type Host interface {
	ID() PeerID
	Context() context.Context
}

type basicHost struct {
	PeerID
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

	h := &basicHost{c: c, PeerID: NewID()}
	if h.h, err = libp2p.New(c, popt...); err != nil {
		return nil, errors.Wrap(err, "libp2p")
	}

	for _, o := range copt {
		if err = o.Apply(h); err != nil {
			break
		}
	}

	return h, err
}

// Context to which the Host is bound
func (h basicHost) Context() context.Context { return h.c }
