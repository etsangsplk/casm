// Package casm implements shared interfaces
package casm

import (
	"github.com/lthibault/casm/pkg/net"
)

// IDer can provide a PeerID
type IDer interface {
	ID() net.PeerID
}

// Addresser can provide an Addr
type Addresser interface {
	Addr() net.Addr
}
