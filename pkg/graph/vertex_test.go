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

func (m mockAddr) Addr() casm.Addr          { return m }
func (mockAddr) MultiAddrs() []ma.Multiaddr { return []ma.Multiaddr{} }
func (mockAddr) Label() casm.HostLabel      { return "test" }

type mockHost struct {
	casm.PeerID
	Ctx context.Context
}

func (h mockHost) Network() casm.Network      { return h }
func (h mockHost) Hook() casm.NetHookManager  { return h }
func (h mockHost) Add(casm.NetHook)           {}
func (h mockHost) Remove(casm.NetHook)        {}
func (h mockHost) Stream() casm.StreamManager { return h }

func (h mockHost) Context() context.Context             { return h.Ctx }
func (h mockHost) Addr() casm.Addr                      { return &mockAddr{casm.NewID()} }
func (h mockHost) PeerAddr(casm.IDer) (casm.Addr, bool) { return &mockAddr{casm.NewID()}, true }

func (mockHost) Register(string, casm.Handler) {}
func (mockHost) Unregister(string)             {}

func (mockHost) Open(context.Context, casm.Addresser, string) (casm.Stream, error) {
	return nil, nil
}

func (mockHost) Connect(context.Context, casm.Addresser) error { return nil }
func (mockHost) Disconnect(casm.IDer)                          {}

func TestVertex(t *testing.T) {
	h := &mockHost{Ctx: context.Background()}

	t.Run("New", func(t *testing.T) {
		t.Run("Succeed", func(t *testing.T) {
			vtx, err := New(h)
			assert.NoError(t, err)
			v := vtx.(*vertex)

			assert.Equal(t, defaultK, v.k)
			assert.Equal(t, defaultL, v.l)
		})

		t.Run("Fail", func(t *testing.T) {
			o := Option(func(*vertex) error {
				return errors.New("")
			})
			_, err := New(h, o)
			assert.Error(t, err)
		})
	})
}
