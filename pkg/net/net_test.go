package net

import (
	"bytes"
	"io"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	b := new(bytes.Buffer)
	s := "hello, world!"

	t.Run("LenHdr", func(t *testing.T) {
		assert.Equal(t, uint16(13), Path(s).lenHdr())
	})

	t.Run("SendTo", func(t *testing.T) {
		assert.NoError(t, Path(s).SendTo(b))
	})

	t.Run("RecvFrom", func(t *testing.T) {
		t.Run("Succeed", func(t *testing.T) {
			defer b.Reset()

			var p Path
			assert.NoError(t, p.RecvFrom(b))
			assert.Equal(t, s, p.String())
		})

		t.Run("FailLen", func(t *testing.T) {
			defer b.Reset()

			var p Path
			assert.Error(t, p.RecvFrom(b))
		})

		t.Run("FailBody", func(t *testing.T) {
			t.Run("EOF", func(t *testing.T) {
				defer b.Reset()
				binary.Write(b, uint16(10))

				var p Path
				assert.Error(t, p.RecvFrom(b))
			})

			t.Run("GenericError", func(t *testing.T) {
				ch := make(chan io.Reader, 1)
				go func() {
					pr, pw := io.Pipe()
					ch <- pr

					binary.Write(pw, uint16(10))
					pw.CloseWithError(errors.New("fail"))
				}()

				var p Path
				assert.EqualError(t, p.RecvFrom(<-ch), "read path: fail")
			})

		})
	})
}

func TestHandlerFunc(t *testing.T) {
	var ok bool
	HandlerFunc(func(*Stream) { ok = true }).Serve(nil)
	assert.True(t, ok)
}
