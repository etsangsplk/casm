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
	var h *basicHost

	c, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("ListenAndServe", func(t *testing.T) {
		h = New(
			OptTransport(transpt),
			OptLogger(log.New(log.OptLevel(log.NullLevel))),
		).(*basicHost)

		assert.NotNil(t, h.l)
		assert.NoError(t, h.ListenAndServe(c, a))
		assert.NotNil(t, h.a)
	})

	t.Run("Network", func(t *testing.T) {

		t.Run("Connect", func(t *testing.T) {
			t.Run("Self", func(t *testing.T) {
				assert.Error(t, h.Network().Connect(c, a))
			})

			t.Run("Existing", func(t *testing.T) {
				conn := mockConn{remote: db}

				h.peers.Add(conn)
				defer h.peers.Del(db)

				assert.Error(t, h.Network().Connect(c, db))
			})

			t.Run("Valid", func(t *testing.T) {
				// TODO
			})
		})

	})

}
