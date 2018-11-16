// Package casm implements raw CASM hosts
package casm

import (
	"math/rand"
	"time"

	"github.com/lthibault/casm/pkg/net"
)

func init() { rand.Seed(time.Now().UTC().UnixNano()) }

// PeerID uniquely identifies a host instance
type PeerID = net.PeerID

// NewID produces a random PeerID
func NewID() PeerID { return PeerID(rand.Uint64()) }

// IDer can provide a PeerID
type IDer interface {
	ID() PeerID
}

// Addresser can provide an Addr
type Addresser interface {
	Addr() net.Addr
}
