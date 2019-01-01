package net

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddr(t *testing.T) {
	a := addr{PeerID: New(), proto: "proto", network: "net", addr: "addr"}
	assert.Equal(t, a, a.Addr())
}

func TestWireAddr(t *testing.T) {
	var wa *wireAddr
	b := new(bytes.Buffer)

	t.Run("FromAddr", func(t *testing.T) {
		a := addr{
			PeerID:  New(),
			proto:   "inproc",
			network: "",
			addr:    "/test",
		}
		wa = newWireAddr(a)
		assertAddrEqual(t, a, wa)
		assertAddrEqual(t, wa, wa.Addr())
	})

	t.Run("SendTo", func(t *testing.T) {
		assert.NoError(t, wa.SendTo(b))
	})

	t.Run("RecvFrom", func(t *testing.T) {
		aw := new(wireAddr)
		assert.NoError(t, aw.RecvFrom(b))
		assertAddrEqual(t, wa, aw)
	})
}
