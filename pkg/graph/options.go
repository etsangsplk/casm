package graph

import casm "github.com/lthibault/casm/pkg"

const (
	defaultK uint8 = 5
	defaultL uint8 = 1
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
			func(v *V) error {
				return casm.OptNetHook(v).(casm.Applicator).Apply(v.h)
			},
		)

		return
	}
}
