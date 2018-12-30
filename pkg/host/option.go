package host

import (
	net "github.com/lthibault/casm/pkg/net"
	log "github.com/lthibault/log/pkg"
)

// Option represents a setting
type Option func(*basicHost) Option

// OptLogger sets the logger
func OptLogger(l log.Logger) Option {
	if l == nil {
		l = log.New()
	}

	return func(bh *basicHost) (prev Option) {
		prev = OptLogger(bh.l)
		bh.l = l
		return
	}
}

// OptTransport sets the net.Transport.
func OptTransport(t *net.Transport) Option {
	return func(bh *basicHost) (prev Option) {
		prev = OptTransport(bh.t)
		bh.t = t
		return
	}
}
