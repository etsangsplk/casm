package casm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeerID(t *testing.T) {
	var pid PeerID

	t.Run("Generate", func(t *testing.T) {
		pid = NewID()
		assert.NotZero(t, pid)
		assert.NotEqual(t, pid, NewID())
		assert.Equal(t, pid, pid.ID())
	})
}

func TestHost(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Run("Succeed", func(t *testing.T) {
			c := context.Background()
			h, err := New(c)
			bh := h.(*basicHost)

			assert.NoError(t, err)
			assert.Equal(t, context.Background(), bh.c)
			assert.NotZero(t, h.Addr().ID())
			assert.Equal(t, c, h.Context())
		})

		// t.Run("Fail", func(t *testing.T) {
		// })
	})

}
