package host

import (
	"context"
	"testing"

	net "github.com/lthibault/casm/pkg/net"
	log "github.com/lthibault/log/pkg"
	"github.com/lthibault/pipewerks/pkg/transport/inproc"
	"github.com/stretchr/testify/assert"
)

var (
	a  = net.NewAddr(net.New(), "", "inproc", "/host")
	db = net.NewAddr(net.New(), "", "inproc", "/dialback")
)

func TestHost(t *testing.T) {
	transpt := net.NewTransport(inproc.New())
	var h *Host

	c, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Start", func(t *testing.T) {
		h = New(
			OptTransport(transpt),
			OptLogger(log.New(log.OptLevel(log.NullLevel))),
		)

		assert.NotNil(t, h.l)
		assert.NoError(t, h.Start(c, a))
		assert.NotNil(t, h.a)
	})

	// t.Run("Network", func(t *testing.T) {

	// })

}
