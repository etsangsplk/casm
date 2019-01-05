package host

import (
	"context"
	"testing"

	net "github.com/lthibault/casm/pkg/net"
	"github.com/stretchr/testify/assert"
)

type mockConn struct {
	remote net.Addr
	closed bool
}

func (mockConn) Context() context.Context              { return context.Background() }
func (mockConn) LocalAddr() net.Addr                   { return nil }
func (c mockConn) RemoteAddr() net.Addr                { return c.remote }
func (mockConn) AcceptStream() (*net.Stream, error)    { return nil, nil }
func (mockConn) OpenStream() (*net.Stream, error)      { return nil, nil }
func (mockConn) WithContext(context.Context) *net.Conn { return nil }
func (c *mockConn) Close() error {
	c.closed = true
	return nil
}

func TestCxnTable(t *testing.T) {
	var ct cxnTable = make(map[net.PeerID]cxn)
	conn := &mockConn{remote: net.NewAddr(net.New(), "", "", "")}

	t.Run("Add", func(t *testing.T) {
		assert.True(t, ct.Add(conn))
	})

	t.Run("Get", func(t *testing.T) {
		c, ok := ct.Get(conn.RemoteAddr().ID())
		assert.True(t, ok)
		assert.Equal(t, conn, c)
	})

	t.Run("Del", func(t *testing.T) {
		c, ok := ct.Del(conn.RemoteAddr().ID())
		assert.True(t, ok)
		assert.Equal(t, conn, c)
	})
}

func TestPeerStore(t *testing.T) {
	p := newPeerStore()
	assert.NotNil(t, p.t)

	conn := &mockConn{remote: net.NewAddr(net.New(), "", "", "")}

	t.Run("StoreOrClose", func(t *testing.T) {
		assert.True(t, p.StoreOrClose(conn))
		assert.False(t, p.StoreOrClose(conn))
		assert.True(t, conn.closed)
		assert.Contains(t, p.t, conn.RemoteAddr().ID())
	})

	t.Run("Retrieve", func(t *testing.T) {
		t.Run("Succeed", func(t *testing.T) {
			c, found := p.Retrieve(conn.RemoteAddr())
			assert.True(t, found)
			assert.Equal(t, conn, c)
		})

		t.Run("Fail", func(t *testing.T) {
			_, found := p.Retrieve(net.New())
			assert.False(t, found)
		})
	})

	t.Run("Contains", func(t *testing.T) {
		t.Run("Succeed", func(t *testing.T) {
			assert.True(t, p.Contains(conn.RemoteAddr()))
		})

		t.Run("Fail", func(t *testing.T) {
			assert.False(t, p.Contains(net.New()))
		})
	})

	t.Run("DropAndClose", func(t *testing.T) {
		p.DropAndClose(conn.RemoteAddr())
		assert.NotContains(t, p.t, conn.RemoteAddr().ID())
	})

	t.Run("Reset", func(t *testing.T) {
		assert.True(t, p.StoreOrClose(conn))
		p.Reset()
		assert.NotContains(t, p.t, conn.RemoteAddr().ID())
	})
}
