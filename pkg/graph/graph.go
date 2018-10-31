// Package graph implements the CASM expander graph model
package graph

import (
	casm "github.com/lthibault/casm/pkg"
)

// Vertex in the expander graph
type Vertex interface {
}

// V is a concrete Vertex
type V struct {
	h    casm.Host
	k, l uint8
}

// New V
func New(h casm.Host, opt ...Option) (v *V, err error) {
	v = &V{h: h}
	for _, o := range append([]Option{OptDefault()}, opt...) {
		if _, err = o(v); err != nil {
			break
		}
	}

	return
}
