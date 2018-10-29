// Package graph implements the CASM expander graph model
package graph

import (
	casm "github.com/lthibault/casm/pkg"
)

// Vertex in the expander graph
type Vertex struct {
	h    casm.Host
	k, l uint8
}

// New Vertex
func New(h casm.Host, opt ...Option) (v *Vertex, err error) {
	v = &Vertex{h: h}
	for _, o := range append([]Option{OptDefault()}, opt...) {
		if _, err = o(v); err != nil {
			break
		}
	}

	return
}
