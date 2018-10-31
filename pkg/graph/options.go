package graph

const (
	defaultK uint8 = 5
	defaultL uint8 = 1
)

// Option type for V
type Option func(*V) (Option, error)

// OptCardinality sets the V cardinality, i.e.: the maximum number of edges
func OptCardinality(k uint8) Option {
	return func(v *V) (prev Option, err error) {
		prev = OptCardinality(v.k)
		v.k = k
		return
	}
}

// OptElasticity sets the V elasticity, i.e.: the number of edges beyond
// its cardinality limit that it can host.
func OptElasticity(l uint8) Option {
	return func(v *V) (prev Option, err error) {
		prev = OptElasticity(v.l)
		v.l = l
		return
	}
}

// OptDefault sets the default options for a V
func OptDefault() Option {
	return func(v *V) (prev Option, err error) {
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
