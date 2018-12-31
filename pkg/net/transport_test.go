package net

import (
	"context"
	"io"
	"testing"

	"github.com/lthibault/pipewerks/pkg/transport/tcp"
	"github.com/stretchr/testify/assert"
)

var pipeTransport = tcp.New()

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

	da := NewAddr(New(), tcpType.String(), "tcp", "localhost:9021")
	la := NewAddr(New(), tcpType.String(), "tcp", "localhost:9022")

	pl := transport.NewListener(la)
	l, err := pl.Listen(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, la, l.Addr())
	// defer assert.NoError(t, l.Close())
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
	defer assertProperClosure(t, dconn)

	assertAddrEqual(t, dconn.LocalAddr(), da)
	assertAddrEqual(t, dconn.RemoteAddr(), la)

	lconn := <-ch
	defer assertProperClosure(t, lconn)
	assertAddrEqual(t, lconn.LocalAddr(), la)
	assertAddrEqual(t, lconn.RemoteAddr(), da)

}

func TestProtoDialer(t *testing.T) {

}

func TestProtoListener(t *testing.T) {

}

func TestListener(t *testing.T) {

}
