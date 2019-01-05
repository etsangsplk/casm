package host

import (
	net "github.com/lthibault/casm/pkg/net"
	log "github.com/lthibault/log/pkg"
	tcp "github.com/lthibault/pipewerks/pkg/transport/tcp"
)

// Option represents a setting
type Option func(*Host) Option

// OptLogger sets the logger
func OptLogger(l log.Logger) Option {
	if l == nil {
		l = log.New()
	}

	return func(h *Host) (prev Option) {
		prev = OptLogger(h.l)
		h.l = l
		return
	}
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

// OptTransport sets the net.Transport.
func OptTransport(t *net.Transport) Option {
	return func(h *Host) (prev Option) {
		prev = OptTransport(h.t)
		h.t = t
		return
	}
}
