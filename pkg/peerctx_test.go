package casm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	peer "github.com/libp2p/go-libp2p-peer"
)

func TestYMap(t *testing.T) {
	ym := newYMap()
	id := NewID()
	hid := peer.ID("test")

	t.Run("Put", func(t *testing.T) {
		ym.Put(context.Background(), id, hid)
		assert.Contains(t, ym.m, id)
		assert.Contains(t, ym.m, hid)
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			c, ok := ym.Get(id)
			assert.True(t, ok)
			assert.Equal(t, id, getID(c))
			assert.Equal(t, hid, getHID(c))
		})

		t.Run("ByHID", func(t *testing.T) {
			c, ok := ym.Get(id)
			assert.True(t, ok)
			assert.Equal(t, id, getID(c))
			assert.Equal(t, hid, getHID(c))
		})
	})

	t.Run("Del", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			ym.Del(id)
			assert.Empty(t, ym.m)
		})

		t.Run("ByHID", func(t *testing.T) {
			ym.Put(context.Background(), id, hid)
			ym.Del(hid)
			assert.Empty(t, ym.m)
		})
	})
}

func TestContextHook(t *testing.T) {

}
