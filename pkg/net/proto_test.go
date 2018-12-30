package net

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/lunixbochs/struc"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockConn struct {
	net.Conn
	t0  time.Time
	err error
}

func (c *mockConn) SetDeadline(t time.Time) error {
	c.t0 = t
	return c.err
}

func TestProto(t *testing.T) {
	buf := new(bytes.Buffer)

	t.Run("WithTimeout", func(t *testing.T) {
		t.Run("NoError", func(t *testing.T) {
			conn := &mockConn{}
			err := proto.withTimeout(conn, func() error {
				t1 := time.Now().Add(upgradeDeadline)
				assert.WithinDuration(t, conn.t0, t1, time.Millisecond)
				return nil
			})
			assert.NoError(t, err)
		})

		t.Run("Error", func(t *testing.T) {
			conn := &mockConn{err: errors.New("")}
			err := proto.withTimeout(conn, func() error {
				t1 := time.Now().Add(upgradeDeadline)
				assert.WithinDuration(t, conn.t0, t1, time.Microsecond)
				return nil
			})
			assert.Error(t, err)
		})
	})

	t.Run("CheckRemoteID", func(t *testing.T) {
		t.Run("Match", func(t *testing.T) {
			defer buf.Reset()
			binary.Write(buf, binary.BigEndian, PeerID(1))
			fn := proto.checkRemoteID(buf, PeerID(1))
			assert.NoError(t, fn())
		})

		t.Run("NoMatch", func(t *testing.T) {
			defer buf.Reset()
			binary.Write(buf, binary.BigEndian, PeerID(1))
			fn := proto.checkRemoteID(buf, PeerID(9))
			assert.Error(t, fn())
		})
	})

	t.Run("RecvDialback", func(t *testing.T) {
		defer buf.Reset()

		a := addr{
			PeerID:  New(),
			proto:   "inproc",
			network: inprocType.String(),
			addr:    "/test",
		}
		assert.NoError(t, struc.Pack(buf, newWireAddr(a)))

		wa := new(wireAddr)
		fn := proto.recvDialback(buf, wa)
		assert.NoError(t, fn())

		assert.Equal(t, a.ID(), wa.ID())
		assert.Equal(t, a.Network(), wa.Network())
		assert.Equal(t, a.Proto(), wa.Proto())
		assert.Equal(t, a.String(), wa.String())
	})
}
