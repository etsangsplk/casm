package net

import (
	"bytes"
	"testing"

	"github.com/lunixbochs/struc"
	"github.com/stretchr/testify/assert"
)

func TestWireAddr(t *testing.T) {
	var wa *wireAddr
	b := new(bytes.Buffer)

	t.Run("FromAddr", func(t *testing.T) {
		a := addr{
			PeerID:  New(),
			proto:   "inproc",
			network: inprocType.String(),
			addr:    "/test",
		}
		wa = newWireAddr(a)
		assert.Equal(t, a.ID(), wa.ID())
		assert.Equal(t, a.Network(), wa.Network())
		assert.Equal(t, a.Proto(), wa.Proto())
		assert.Equal(t, a.String(), wa.String())
	})

	t.Run("MarshalBinary", func(t *testing.T) {
		assert.NoError(t, struc.Pack(b, wa))
	})

	t.Run("UnmarshalBinary", func(t *testing.T) {
		var aw wireAddr
		assert.NoError(t, struc.Unpack(b, &aw))
		assert.Equal(t, aw.ID(), wa.ID())
		assert.Equal(t, aw.Network(), wa.Network())
		assert.Equal(t, aw.Proto(), wa.Proto())
		assert.Equal(t, aw.String(), wa.String())
	})
}
