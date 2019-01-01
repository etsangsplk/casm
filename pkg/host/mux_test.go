package host

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"testing"

	log "github.com/lthibault/log/pkg"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type testHandler bool

func (h *testHandler) Serve(Stream) { *h = true }

func TestPath(t *testing.T) {
	b := new(bytes.Buffer)
	s := "hello, world!"

	t.Run("LenHdr", func(t *testing.T) {
		assert.Equal(t, uint16(13), path(s).lenHdr())
	})

	t.Run("SendTo", func(t *testing.T) {
		assert.NoError(t, path(s).SendTo(b))
	})

	t.Run("RecvFrom", func(t *testing.T) {
		t.Run("Succeed", func(t *testing.T) {
			defer b.Reset()

			var p path
			assert.NoError(t, p.RecvFrom(b))
			assert.Equal(t, s, p.String())
		})

		t.Run("FailLen", func(t *testing.T) {
			defer b.Reset()

			var p path
			assert.Error(t, p.RecvFrom(b))
		})

		t.Run("FailBody", func(t *testing.T) {
			t.Run("EOF", func(t *testing.T) {
				defer b.Reset()
				binary.Write(b, binary.BigEndian, uint16(10))

				var p path
				assert.Error(t, p.RecvFrom(b))
			})

			t.Run("GenericError", func(t *testing.T) {
				ch := make(chan io.Reader, 1)
				go func() {
					pr, pw := io.Pipe()
					ch <- pr

					binary.Write(pw, binary.BigEndian, uint16(10))
					pw.CloseWithError(errors.New("fail"))
				}()

				var p path
				assert.EqualError(t, p.RecvFrom(<-ch), "read path: fail")
			})

		})
	})
}

func TestHandlerFunc(t *testing.T) {
	var ok bool
	HandlerFunc(func(Stream) { ok = true }).Serve(nil)
	assert.True(t, ok)
}

func TestMux(t *testing.T) {
	var wg sync.WaitGroup
	m := newStreamMux(log.New(log.OptLevel(log.NullLevel))) // disable logging

	t.Run("Register", func(t *testing.T) {
		wg.Add(100)

		for i := 0; i < 100; i++ {
			go func(n int) {
				defer wg.Done()
				m.Register(fmt.Sprintf("/test/%d", n), new(testHandler))
			}(i)
		}

		wg.Wait()

		assert.Equal(t, 100, m.r.Len())
	})

	t.Run("Serve", func(t *testing.T) {
		wg.Add(100)

		for i := 0; i < 100; i++ {
			go func(n int) {
				defer wg.Done()
				m.Serve(stream{path: fmt.Sprintf("/test/%d", n)})
			}(i)
		}

		wg.Wait()

		m.r.Walk(func(s string, v interface{}) (terminate bool) {
			assert.True(t, bool(*(v.(*testHandler))), fmt.Sprintf("%s not called", s))
			return
		})
	})

	t.Run("Replace", func(t *testing.T) {
		// N.B.: this test enforces a detail of the mux spec; Registering a Handler
		// 		 to an already-registered path _replaces_ the existing Handler.
		m.Register("/test/0", new(testHandler))

		v, ok := m.r.Get("/test/0")
		assert.True(t, ok, "/test/0 was removed")
		assert.False(t, bool(*(v.(*testHandler))), "/test/0 was called or not replaced")
	})

	t.Run("Unregister", func(t *testing.T) {
		wg.Add(100)

		for i := 0; i < 100; i++ {
			go func(n int) {
				defer wg.Done()
				m.Unregister(fmt.Sprintf("/test/%d", n))
			}(i)
		}

		wg.Wait()

		assert.Zero(t, m.r.Len())
	})
}
