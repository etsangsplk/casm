package graph

import (
	casm "github.com/lthibault/casm/pkg"
)

// type

// peerstore.PeerInfo

// Neighborhood is a view of peers adjancent to a given Vertex.
type Neighborhood interface {
	Lease()
	Evict(casm.PeerID)
}
