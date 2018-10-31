package graph

import (
	"testing"

	casm "github.com/lthibault/casm/pkg"
	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {
	id := casm.NewID()

	t.Run("Factory", func(t *testing.T) {
		f := newMsgFactory(id)

		t.Run("FirstSequenceIsOne", func(t *testing.T) {
			msg := f(nil)
			defer msg.Free()
			assert.Equal(t, uint64(1), msg.Sequence())
		})

		t.Run("PeerID", func(t *testing.T) {
			msg := f(nil)
			defer msg.Free()
			assert.Equal(t, id, msg.ID())
		})

		t.Run("RefCounter", func(t *testing.T) {
			msg := f(nil)
			t.Run("IsSet", func(t *testing.T) {
				assert.Equal(t, uint32(1), msg.ctr)
			})

			t.Run("Incr", func(t *testing.T) {
				msg.Ref()
				assert.Equal(t, uint32(2), msg.ctr)
			})

			t.Run("Decr", func(t *testing.T) {
				msg.Free()
				assert.Equal(t, uint32(1), msg.ctr)
			})

			t.Run("Free", func(t *testing.T) {
				msg.Free()
				assert.Zero(t, msg.ctr)
			})
		})

		t.Run("BodyIsSet", func(t *testing.T) {
			msg := f([]byte("body"))
			defer msg.Free()
			assert.Equal(t, []byte("body"), msg.Body())
		})
	})

}
