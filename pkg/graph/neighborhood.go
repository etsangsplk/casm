package graph

import (
	"context"

	casm "github.com/lthibault/casm/pkg"
)

// Neighborhood is a view of peers adjancent to a given Vertex.
type Neighborhood interface {
	Lease(context.Context, casm.Addr) error
	Evict(casm.IDer)
}
