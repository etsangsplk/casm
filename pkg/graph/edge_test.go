package graph

import (
	"context"
	"io"
	"testing"
	"time"

	casm "github.com/lthibault/casm/pkg"
	"github.com/stretchr/testify/assert"
)

var _ casm.Stream = &mockStream{} // type-constraint

type mockStream struct {
	c      context.Context
	cancel func()
	io.ReadWriter
	rpid casm.IDer
}

func newStream(rw io.ReadWriter) *mockStream {
	c, cancel := context.WithCancel(context.Background())
	return &mockStream{
		c:          c,
		cancel:     cancel,
		ReadWriter: rw,
		rpid:       casm.NewID(),
	}
}

func (m mockStream) Context() context.Context { return m.c }
func (m mockStream) CloseWrite() error        { return nil }
func (m mockStream) Close() error {
	m.cancel()
	return nil
}
func (m mockStream) RemotePeer() casm.PeerID          { return m.rpid.ID() }
func (m mockStream) SetDeadline(time.Time) error      { return nil }
func (m mockStream) SetReadDeadline(time.Time) error  { return nil }
func (m mockStream) SetWriteDeadline(time.Time) error { return nil }

func TestEdgeNegotiator(t *testing.T) {
	t.Run("maybeInitUnsafe", func(t *testing.T) {
		id := casm.NewID()
		en := newEdgeNegotiator()
		assert.Empty(t, en.m)
		assert.NotNil(t, en.maybeInitUnsafe(id))
		assert.NotEmpty(t, en.m)
	})

	t.Run("ProvideBeforeWait", func(t *testing.T) {
		id := casm.NewID()
		en := newEdgeNegotiator()
		assert.NotNil(t, en.ProvideDataStream(id))
		assert.NotNil(t, en.WaitDataStream(id))

		en.Clear(id)
		assert.Empty(t, en.m)
	})

	t.Run("WaitBeforeProvide", func(t *testing.T) {
		id := casm.NewID()
		en := newEdgeNegotiator()
		assert.NotNil(t, en.ProvideDataStream(id))
		assert.NotNil(t, en.WaitDataStream(id))

		en.Clear(id)
		assert.Empty(t, en.m)
	})

	t.Run("Send", func(t *testing.T) {
		t.Run("ProvideBeforeWait", func(t *testing.T) {
			id := casm.NewID()
			s := &mockStream{}
			en := newEdgeNegotiator()

			go func() { en.ProvideDataStream(id) <- s }()
			assert.Equal(t, s, <-en.WaitDataStream(id))
		})

		t.Run("WaitBeforeProvide", func(t *testing.T) {
			id := casm.NewID()
			s := &mockStream{}
			en := newEdgeNegotiator()

			ch := en.WaitDataStream(id)
			go func() { en.ProvideDataStream(id) <- s }()
			assert.Equal(t, s, <-ch)
		})
	})
}
