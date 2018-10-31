package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	v := new(V)

	t.Run("Cardinality", func(t *testing.T) {
		var prev Option
		var err error
		orig := v.k

		t.Run("Set", func(t *testing.T) {
			prev, err = OptCardinality(defaultK)(v)
			assert.NoError(t, err)
			assert.Equal(t, defaultK, v.k)
		})

		t.Run("Prev", func(t *testing.T) {
			_, err = prev(v)
			assert.NoError(t, err)
			assert.Equal(t, orig, v.k)
		})
	})

	t.Run("Elasticity", func(t *testing.T) {
		var prev Option
		var err error
		orig := v.l

		t.Run("Set", func(t *testing.T) {
			prev, err = OptElasticity(defaultL)(v)
			assert.NoError(t, err)
			assert.Equal(t, defaultL, v.l)
		})

		t.Run("Prev", func(t *testing.T) {
			_, err = prev(v)
			assert.NoError(t, err)
			assert.Equal(t, orig, v.l)
		})
	})
}
