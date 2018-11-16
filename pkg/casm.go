// Package casm implements raw CASM hosts
package casm

import (
	"math/rand"
	"time"

	"github.com/lthibault/casm/pkg/net"
)

func init() { rand.Seed(time.Now().UTC().UnixNano()) }

type (
	// PeerID uniquely identifies a host instance
	PeerID = net.PeerID

	// IDer can provide a PeerID
	IDer = net.IDer
)

// NewID produces a random PeerID
func NewID() PeerID { return PeerID(rand.Uint64()) }

// Addresser can provide an Addr
type Addresser interface {
	Addr() net.Addr
}
