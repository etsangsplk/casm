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
	"golang.org/x/sync/errgroup"
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
			err := withTimeout(conn, func() error {
				t1 := time.Now().Add(upgradeDeadline)
				assert.WithinDuration(t, conn.t0, t1, time.Millisecond)
				return nil
			})
			assert.NoError(t, err)
		})

		t.Run("Error", func(t *testing.T) {
			conn := &mockConn{err: errors.New("")}
			err := withTimeout(conn, func() error {
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
			fn := checkRemoteID(buf, PeerID(1))
			assert.NoError(t, fn())
		})

		t.Run("NoMatch", func(t *testing.T) {
			defer buf.Reset()
			binary.Write(buf, binary.BigEndian, PeerID(1))
			fn := checkRemoteID(buf, PeerID(9))
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
		fn := recvDialback(buf, wa)
		assert.NoError(t, fn())

		assert.Equal(t, a.ID(), wa.ID())
		assert.Equal(t, a.Network(), wa.Network())
		assert.Equal(t, a.Proto(), wa.Proto())
		assert.Equal(t, a.String(), wa.String())
	})

	t.Run("SendID", func(t *testing.T) {
		defer buf.Reset()

		fn := sendID(buf, PeerID(1))
		assert.NoError(t, fn())

		var pid PeerID
		assert.NoError(t, binary.Read(buf, binary.BigEndian, &pid))
		assert.Equal(t, PeerID(1), pid)
	})

	t.Run("SendDialback", func(t *testing.T) {
		defer buf.Reset()

		a := addr{
			PeerID:  New(),
			proto:   "inproc",
			network: inprocType.String(),
			addr:    "/test",
		}

		fn := sendDialback(buf, a)
		assert.NoError(t, fn())

		wa := new(wireAddr)
		assert.NoError(t, struc.Unpack(buf, wa))

		assert.Equal(t, a.ID(), wa.ID())
		assert.Equal(t, a.Network(), wa.Network())
		assert.Equal(t, a.Proto(), wa.Proto())
		assert.Equal(t, a.String(), wa.String())
	})

	t.Run("Integration", func(t *testing.T) {
		dc, lc := net.Pipe()

		da := addr{
			PeerID:  New(),
			proto:   "inproc",
			network: inprocType.String(),
			addr:    "/test/alpha",
		}

		la := addr{
			PeerID:  New(),
			proto:   "inproc",
			network: inprocType.String(),
			addr:    "/test/bravo",
		}

		var a Addr
		var g errgroup.Group
		g.Go(func() error {
			return proto.upgradeDialer(dc, da, la.ID())
		})
		g.Go(func() (err error) {
			a, err = proto.upgradeListener(lc, la)
			return err
		})
		assert.NoError(t, g.Wait())

		assert.Equal(t, da.ID(), a.ID())
		assert.Equal(t, da.Network(), a.Network())
		assert.Equal(t, da.Proto(), a.Proto())
		assert.Equal(t, da.String(), a.String())
	})
}
