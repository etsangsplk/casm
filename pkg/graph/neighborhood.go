package graph

import (
	"context"

	casm "github.com/lthibault/casm/pkg"
)

// Neighborhood is a view of peers adjancent to a given Vertex.
type Neighborhood interface {
	Connected(casm.IDer) bool
	Lease(context.Context, casm.Addresser) error
	Evict(casm.IDer)
}
