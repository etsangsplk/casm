package host

import (
	casm "github.com/lthibault/casm/pkg"
	net "github.com/lthibault/casm/pkg/net"
	"github.com/lthibault/casm/pkg/transport/quic"
)

type ctxKey uint16

const (
	keyListenAddr ctxKey = iota
)

// Option represents a setting
type Option func(*basicHost) Option

// OptListenAddr sets the listen address
func OptListenAddr(addr string) Option {
	return func(h *basicHost) (prev Option) {
		h.addr = addr
		return
	}
}

// OptTransport sets the transport
func OptTransport(t net.Transport) Option {
	return func(h *basicHost) (prev Option) {
		h.t = t
		return
	}
}

func optSetID() Option {
	return func(h *basicHost) (prev Option) {
		prev = optSetID()
		h.id = casm.NewID()
		return
	}
}

func defaultQUIC() *quic.Config {
	return &quic.Config{
		KeepAlive: true,
	}
}

func maybeMkQUIC() Option {
	return func(h *basicHost) (prev Option) {
		prev = OptTransport(h.t)
		if h.t == nil {
			tc := generateTLSConfig()
			qc := defaultQUIC()

			h.t = quic.NewTransport(h.id, quic.OptQuic(qc), quic.OptTLS(tc))
		}
		return
	}
}

func defaultOpts(overrides ...Option) []Option {
	opt := []Option{
		optSetID(),
		OptListenAddr("localhost:1987"),
	}

	return append(opt, maybeMkQUIC())
}
