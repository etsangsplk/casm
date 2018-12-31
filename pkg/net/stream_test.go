package net

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockAddrs struct{ local, remote Addr }

func (a mockAddrs) LocalAddr() Addr  { return a.local }
func (a mockAddrs) RemoteAddr() Addr { return a.remote }

func TestStream(t *testing.T) {
	var s Stream
	t.Run("WithContext", func(t *testing.T) {
		c := context.Background()
		assert.Equal(t, c, s.WithContext(c).Context())
	})
}
