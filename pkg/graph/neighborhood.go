package graph

import (
	casm "github.com/lthibault/casm/pkg"
)

// Neighborhood is a view of peers adjancent to a given Vertex.
type Neighborhood interface {
	Lease(casm.Addr)
	Evict(casm.IDer)
}
