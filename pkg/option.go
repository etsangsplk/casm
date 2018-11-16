package casm

import (
	"github.com/lthibault/casm/pkg/transport/quic"
)

// Option represents a setting
type Option func(*basicHost) Option

// OptListenAddr sets the listen address
func OptListenAddr(addr string) Option {
	return func(h *basicHost) (prev Option) {
		h.addr.addr = addr
		return
	}
}

// OptTransport sets the transport
func OptTransport(t Transport) Option {
	return func(h *basicHost) (prev Option) {
		h.t = t
		return
	}
}

func optSetID() Option {
	return func(h *basicHost) (prev Option) {
		prev = optSetID()
		h.addr.IDer = NewID()
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

			h.t = quic.NewTransport(quic.OptQuic(qc), quic.OptTLS(tc))
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
