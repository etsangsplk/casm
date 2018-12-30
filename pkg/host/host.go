package host

import (
	"context"

	"github.com/lthibault/pipewerks/pkg/transport/generic"
	"github.com/lthibault/pipewerks/pkg/transport/tcp"

	casm "github.com/lthibault/casm/pkg"
	net "github.com/lthibault/casm/pkg/net"
)

// Network manages raw connections
type Network interface {
	Connect(context.Context, casm.Addresser) error
	Disconnect(casm.IDer)
}

// StreamManager manages streams, which are multiplexed on top of raw connections
type StreamManager interface {
	Register(string, net.Handler)
	Unregister(string)
	Open(casm.Addresser, string) (*net.Stream, error)
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

func defaultTransportFactory() *net.Transport {
	return net.NewTransport(
		tcp.New(tcp.OptGeneric(generic.OptConnectHandler(net.RawConnUpgrader{}))),
		net.RawConnUpgrader{},
	)
}

func setDefaultOpts(opt []Option) []Option {
	return append(
		[]Option{
			OptTransport(defaultTransportFactory()),
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

	bh.mux = newStreamMux(bh.log().WithLocus("mux"))
	bh.peers = newPeerStore()
	return bh
}
