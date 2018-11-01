package graph

const (
	defaultK uint8 = 5
	defaultL uint8 = 1

	pathEdge = "/edge"
)

// Option type for V
type Option func(*V) error

// OptCardinality sets the V cardinality, i.e.: the maximum number of edges
func OptCardinality(k uint8) Option {
	return func(v *V) (err error) {
		v.k = k
		return
	}
}

// OptElasticity sets the V elasticity, i.e.: the number of edges beyond
// its cardinality limit that it can host.
func OptElasticity(l uint8) Option {
	return func(v *V) (err error) {
		v.l = l
		return
	}
}

// // OptExtensions registers extensions to the Vertex
// func OptExtensions(ext ...casm.Handler) Option {
// 	return func(v *V) (err error) {

// 	}
// }

// OptDefault sets the default options for a V
func OptDefault() Option {
	return func(v *V) (err error) {
		apply := func(opt ...Option) {
			for _, o := range opt {
				if err == nil {
					err = o(v)
				}
			}
		}

		apply(
			OptCardinality(defaultK),
			OptElasticity(defaultL),
		)

		return
	}
}
