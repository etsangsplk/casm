package graph

const (
	defaultK uint8 = 5
	defaultL uint8 = 1
)

// Option type for Vertex
type Option func(*Vertex) (Option, error)

// OptCardinality sets the Vertex cardinality, i.e.: the maximum number of edges
func OptCardinality(k uint8) Option {
	return func(v *Vertex) (prev Option, err error) {
		prev = OptCardinality(v.k)
		v.k = k
		return
	}
}

// OptElasticity sets the Vertex elasticity, i.e.: the number of edges beyond
// its cardinality limit that it can host.
func OptElasticity(l uint8) Option {
	return func(v *Vertex) (prev Option, err error) {
		prev = OptElasticity(v.l)
		v.l = l
		return
	}
}

// OptDefault sets the default options for a Vertex
func OptDefault() Option {
	return func(v *Vertex) (prev Option, err error) {
		apply := func(opt ...Option) {
			for _, o := range opt {
				if err == nil {
					prev, err = o(v)
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
