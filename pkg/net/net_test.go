package net

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	b := new(bytes.Buffer)
	s := "hello, world!"

	t.Run("LenHdr", func(t *testing.T) {
		assert.Equal(t, uint16(13), Path(s).lenHdr())
	})

	t.Run("WriteTo", func(t *testing.T) {
		assert.NoError(t, Path(s).WriteTo(b))
	})

	t.Run("ReadFrom", func(t *testing.T) {
		var p Path
		assert.NoError(t, p.ReadFrom(b))
		assert.Equal(t, s, p.String())
	})
}
