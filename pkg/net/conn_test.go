package net

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConn(t *testing.T) {
	var conn Conn
	t.Run("WithContext", func(t *testing.T) {
		c := context.Background()
		assert.Equal(t, c, conn.WithContext(c).Context())
	})
}
