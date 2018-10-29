package casm

import (
	"context"
	"testing"

	"github.com/libp2p/go-libp2p/config"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestPeerID(t *testing.T) {
	var pid PeerID

	t.Run("Generate", func(t *testing.T) {
		pid = NewID()
		assert.NotZero(t, pid)
		assert.NotEqual(t, pid, NewID())
		assert.Equal(t, pid, pid.ID())
	})

	t.Run("LoadHex", func(t *testing.T) {
		pidPrime, err := IDFromHex(pid.String())
		assert.NoError(t, err)
		assert.Equal(t, pid, pidPrime)
	})
}

func TestHost(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Run("Succeed", func(t *testing.T) {
			c := context.Background()
			h, err := New(c)
			bh := h.(*basicHost)

			assert.NoError(t, err)
			assert.Equal(t, context.Background(), bh.c)
			assert.NotZero(t, h.ID)
			assert.Equal(t, c, h.Context())
		})

		t.Run("Fail", func(t *testing.T) {

			t.Run("libp2pOpt", func(t *testing.T) {
				errOpt := p2pOpt(func(*config.Config) error {
					return errors.New("TESTING ERROR")
				})
				_, err := New(context.Background(), errOpt)
				assert.Error(t, err)
				assert.NotNil(t, errors.Cause(err))
			})

			t.Run("CASMOpt", func(t *testing.T) {
				errOpt := hostOpt(func(*basicHost) error {
					return errors.New("TESTING ERROR")
				})
				_, err := New(context.Background(), errOpt)
				assert.Error(t, err)
			})
		})
	})

}
