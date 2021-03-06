package net

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStream(t *testing.T) {
	var s Stream
	t.Run("WithContext", func(t *testing.T) {
		c := context.Background()
		assert.Equal(t, c, s.WithContext(c).Context())
	})
}
