package casm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHost(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		h, err := New(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, context.Background(), h.c)
		assert.NotZero(t, h.ID)
	})
}
