package graph

import (
	"context"
	"errors"
	"testing"

	casm "github.com/lthibault/casm/pkg"
	"github.com/stretchr/testify/assert"

	ma "github.com/multiformats/go-multiaddr"
)

type mockAddr struct{ casm.IDer }

func (mockAddr) Addrs() []ma.Multiaddr { return []ma.Multiaddr{} }
func (mockAddr) Label() casm.HostLabel { return "test" }

type mockHost struct {
	casm.PeerID
	Ctx context.Context
}

func (h mockHost) Context() context.Context { return h.Ctx }
func (h mockHost) Addr() casm.Addr          { return &mockAddr{casm.NewID()} }

func (mockHost) RegisterStreamHandler(string, casm.Handler) {}
func (mockHost) UnregisterStreamHandler(string)             {}

func (mockHost) OpenStream(context.Context, casm.Addr, string) casm.Stream {
	panic("OpenStream NOT IMPLEMENTED")
}

func TestVertex(t *testing.T) {
	h := &mockHost{Ctx: context.Background()}

	t.Run("New", func(t *testing.T) {
		t.Run("Succeed", func(t *testing.T) {
			v, err := New(h)
			assert.NoError(t, err)
			assert.Equal(t, defaultK, v.k)
			assert.Equal(t, defaultL, v.l)
		})

		t.Run("Fail", func(t *testing.T) {
			o := Option(func(*V) error {
				return errors.New("")
			})
			_, err := New(h, o)
			assert.Error(t, err)
		})
	})
}
