package host

import (
	"context"
	"testing"

	net "github.com/lthibault/casm/pkg/net"
	"github.com/stretchr/testify/assert"
)

type mockConn struct{ remote net.Addr }

func (mockConn) Context() context.Context              { return context.Background() }
func (mockConn) Close() error                          { return nil }
func (mockConn) LocalAddr() net.Addr                   { return nil }
func (c mockConn) RemoteAddr() net.Addr                { return c.remote }
func (mockConn) AcceptStream() (*net.Stream, error)    { return nil, nil }
func (mockConn) OpenStream() (*net.Stream, error)      { return nil, nil }
func (mockConn) WithContext(context.Context) *net.Conn { return nil }

func TestPeerStore(t *testing.T) {
	var p peerMap = make(map[net.PeerID]cxn)
	conn := mockConn{remote: net.NewAddr(net.New(), "", "", "")}

	t.Run("Add", func(t *testing.T) {
		assert.True(t, p.Add(conn))
	})

	t.Run("Get", func(t *testing.T) {
		c, ok := p.Get(conn.RemoteAddr().ID())
		assert.True(t, ok)
		assert.Equal(t, conn, c)
	})

	t.Run("Del", func(t *testing.T) {
		c, ok := p.Del(conn.RemoteAddr().ID())
		assert.True(t, ok)
		assert.Equal(t, conn, c)
	})
}
