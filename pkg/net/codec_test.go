package net

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBinaryCodec(t *testing.T) {
	pid := New()
	b := new(bytes.Buffer)

	t.Run("Write", func(t *testing.T) {
		assert.NoError(t, binary.Write(b, pid))
	})

	t.Run("Read", func(t *testing.T) {
		var p PeerID
		assert.NoError(t, binary.Read(b, &p))
		assert.Equal(t, pid, p)
	})
}
