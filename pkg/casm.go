// Package casm implements shared interfaces
package casm

import (
	"context"

	"github.com/lthibault/casm/pkg/net"
)

// Contexter has a context associated with it
type Contexter interface {
	Context() context.Context
}

// IDer can provide a PeerID
type IDer interface {
	ID() net.PeerID
}

// Addresser can provide an Addr
type Addresser interface {
	Addr() net.Addr
}
