package graph

const (
	defaultK uint8 = 5
	defaultL uint8 = 1
)

// Option type for V
type Option func(*vertex) error

// OptCardinality sets the V cardinality, i.e.: the maximum number of edges
func OptCardinality(k uint8) Option {
	return func(v *vertex) (err error) {
		v.k = k
		return
	}
}

// OptElasticity sets the V elasticity, i.e.: the number of edges beyond
// its cardinality limit that it can host.
func OptElasticity(l uint8) Option {
	return func(v *vertex) (err error) {
		v.l = l
		return
	}
}

// OptDefault sets the default options for a V
func OptDefault() Option {
	return func(v *vertex) (err error) {
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
			func(v *vertex) error {
				v.h.Network().Hook().Add(v)
				return nil
			},
		)

		return
	}
}
