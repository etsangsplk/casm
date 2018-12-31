package net

import (
	"context"
	"encoding/binary"
	"io"
	"testing"

	"github.com/lthibault/pipewerks/pkg/transport/inproc"
	"github.com/stretchr/testify/assert"
)

var pipeTransport = inproc.New()

func assertAddrEqual(t *testing.T, expected, actual Addr) {
	assert.Equal(t, expected.ID(), actual.ID())
	assert.Equal(t, expected.Network(), actual.Network())
	assert.Equal(t, expected.Proto(), actual.Proto())
	assert.Equal(t, expected.String(), actual.String())
}

func assertProperClosure(t *testing.T, c io.Closer) {
	assert.NoError(t, c.Close())
}

func TestTransport(t *testing.T) {
	transport := NewTransport(pipeTransport)

	da := NewAddr(New(), inprocType.String(), "inproc", "/test/dialer")
	la := NewAddr(New(), inprocType.String(), "inproc", "/test/listener")

	pl := transport.NewListener(la)
	l, err := pl.Listen(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, la, l.Addr())
	defer assertProperClosure(t, l)

	ch := make(chan *Conn)
	go func() {
		conn, err := l.Accept()
		assert.NoError(t, err)
		assertAddrEqual(t, conn.LocalAddr(), la)
		assertAddrEqual(t, conn.RemoteAddr(), da)
		ch <- conn
	}()

	pd := transport.NewDialer(da)
	dconn, err := pd.Dial(context.Background(), la)
	assert.NoError(t, err)
	// defer assertProperClosure(t, dconn)  // TODO:  causes panic in YAMUX ...

	assertAddrEqual(t, dconn.LocalAddr(), da)
	assertAddrEqual(t, dconn.RemoteAddr(), la)

	lconn := <-ch
	// defer assertProperClosure(t, lconn)  // TODO:  causes panic in YAMUX ...
	assertAddrEqual(t, lconn.LocalAddr(), la)
	assertAddrEqual(t, lconn.RemoteAddr(), da)

	t.Run("Stream", func(t *testing.T) {
		t.Run("DialerOpens", func(t *testing.T) {
			t.Parallel()
			t.Run("OpenStream", func(t *testing.T) {
				s, err := dconn.OpenStream()
				assert.NoError(t, err)
				defer assertProperClosure(t, s)

				assertAddrEqual(t, s.LocalAddr(), da)
				assertAddrEqual(t, s.RemoteAddr(), la)
			})

			t.Run("AcceptStream", func(t *testing.T) {
				s, err := lconn.AcceptStream()
				assert.NoError(t, err)
				defer assertProperClosure(t, s)

				assertAddrEqual(t, s.LocalAddr(), la)
				assertAddrEqual(t, s.RemoteAddr(), da)
			})
		})

		t.Run("ListenerOpens", func(t *testing.T) {
			t.Parallel()

			t.Run("OpenStream", func(t *testing.T) {
				s, err := lconn.OpenStream()
				assert.NoError(t, err)
				defer assertProperClosure(t, s)

				assertAddrEqual(t, s.LocalAddr(), la)
				assertAddrEqual(t, s.RemoteAddr(), da)
			})

			t.Run("AcceptStream", func(t *testing.T) {
				s, err := dconn.AcceptStream()
				assert.NoError(t, err)
				defer assertProperClosure(t, s)

				assertAddrEqual(t, s.LocalAddr(), da)
				assertAddrEqual(t, s.RemoteAddr(), la)
			})

		})
	})

	t.Run("ReadWrite", func(t *testing.T) {
		t.Parallel()
		t.Run("DialerStream", func(t *testing.T) {
			t.Parallel()

			s, err := lconn.OpenStream()
			assert.NoError(t, err)
			defer assertProperClosure(t, s)

			t.Run("Send", func(t *testing.T) {
				assert.NoError(t, binary.Write(s, binary.BigEndian, uint8(255)))
			})
			t.Run("Recv", func(t *testing.T) {
				var res uint8
				assert.NoError(t, binary.Read(s, binary.BigEndian, &res))
				assert.Equal(t, uint8(127), res)
			})
		})

		t.Run("ListenerStream", func(t *testing.T) {
			t.Parallel()

			s, err := dconn.AcceptStream()
			assert.NoError(t, err)
			defer assertProperClosure(t, s)

			t.Run("Send", func(t *testing.T) {
				assert.NoError(t, binary.Write(s, binary.BigEndian, uint8(127)))
			})
			t.Run("Recv", func(t *testing.T) {
				var res uint8
				assert.NoError(t, binary.Read(s, binary.BigEndian, &res))
				assert.Equal(t, uint8(255), res)
			})
		})
	})
}

func TestProtoDialer(t *testing.T) {

}

func TestProtoListener(t *testing.T) {

}

func TestListener(t *testing.T) {

}
