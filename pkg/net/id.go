package net

import (
	"fmt"
	"math/rand"
	"time"
)

func init() { rand.Seed(time.Now().UTC().UnixNano()) }

// PeerID is a unique identifier for a Node
type PeerID uint64

// New instance
func New() PeerID { return PeerID(rand.Uint64()) }

func (id PeerID) String() string { return fmt.Sprintf("%016x", uint64(id)) }

// ID satisfies the IDer interface
func (id PeerID) ID() PeerID { return id }
