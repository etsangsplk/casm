// Package host implements CASM's p2p host model
package host

import (
	"context"
	"time"

	tcp "github.com/lthibault/pipewerks/pkg/transport/tcp"

	casm "github.com/lthibault/casm/pkg"
	net "github.com/lthibault/casm/pkg/net"
)

// Stream is a full-duplex stream on top of a single physical connection to
// a remote host
type Stream interface {
	Path() string
	Context() context.Context
	StreamID() uint32
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close() error
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// Network manages raw connections
type Network interface {
	Connect(context.Context, casm.Addresser) error
	Disconnect(casm.IDer)
}

// StreamManager manages streams, which are multiplexed on top of raw connections
type StreamManager interface {
	Register(string, Handler)
	Unregister(string)
	Open(casm.Addresser, string) (Stream, error)
}

// Host is a logical machine in a compute cluster.  It acts both as a server and
// a client.  In the CASM expander-graph model, it is a vertex.
type Host interface {
	Context() context.Context
	Addr() net.Addr
	Network() Network
	Stream() StreamManager
	ListenAndServe(context.Context, net.Addr) error
}

func setDefaultOpts(opt []Option) []Option {
	return append(
		[]Option{
			OptTransport(net.NewTransport(tcp.New())),
			OptLogger(nil),
		},
		opt...,
	)
}

// New Host.  Pass options to override defaults.
func New(opt ...Option) Host {
	bh := new(basicHost)

	for _, fn := range setDefaultOpts(opt) {
		fn(bh)
	}

	bh.streamMux = newStreamMux(bh.log().WithLocus("mux"))
	bh.peers = newPeerStore()
	return bh
}
