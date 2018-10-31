package graph

import (
	"context"
	"errors"
	"testing"

	casm "github.com/lthibault/casm/pkg"
	"github.com/stretchr/testify/assert"
)

type mockHost struct {
	casm.PeerID
	Ctx context.Context
}

func (h mockHost) Context() context.Context { return h.Ctx }
func (h mockHost) Addr() casm.Addr          { return nil }

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
			o := Option(func(*V) (Option, error) {
				return nil, errors.New("")
			})
			_, err := New(h, o)
			assert.Error(t, err)
		})
	})
}
