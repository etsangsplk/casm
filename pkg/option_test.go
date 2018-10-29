package casm

import (
	"testing"

	"github.com/libp2p/go-libp2p/config"
	"github.com/stretchr/testify/assert"
)

func TestOptionLoaders(t *testing.T) {
	t.Run("libp2p", func(t *testing.T) {
		var opt p2pOptions
		o := p2pOpt(func(*config.Config) error { return nil })

		opt.Load([]Option{o})
		assert.NotEmpty(t, opt)
	})

	t.Run("casm", func(t *testing.T) {
		var opt hostOptions
		o := hostOpt(func(*basicHost) error { return nil })

		opt.Load([]Option{o})
		assert.NotEmpty(t, opt)
	})
}
