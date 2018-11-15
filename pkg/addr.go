package casm

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"
)

func init() { rand.Seed(time.Now().UTC().UnixNano()) }

// IDer can provide a PeerID
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

// HostLabel identifies a physical host.  You almost certainly don't want this.
// Use PeerID instead.
type HostLabel string

// Addresser can provide an Addr
type Addresser interface {
	Addr() Addr
}

// Addr of a Host
type Addr interface {
	IDer
	Addr() Addr
	net.Addr
}

type addr struct {
	IDer
	addr string
}

func (a addr) Addr() Addr      { return a }
func (a addr) Network() string { return "udp" }
func (a addr) String() string  { return a.addr }
