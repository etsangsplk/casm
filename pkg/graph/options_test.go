package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	v := new(V)

	t.Run("Cardinality", func(t *testing.T) {
		t.Run("Set", func(t *testing.T) {
			err := OptCardinality(defaultK)(v)
			assert.NoError(t, err)
			assert.Equal(t, defaultK, v.k)
		})
	})

	t.Run("Elasticity", func(t *testing.T) {
		t.Run("Set", func(t *testing.T) {
			err := OptElasticity(defaultL)(v)
			assert.NoError(t, err)
			assert.Equal(t, defaultL, v.l)
		})
	})
}
