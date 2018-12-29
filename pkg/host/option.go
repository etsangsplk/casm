package host

import (
	"context"

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
		var old log.Logger

		if bh.c == nil {
			old = nil
			bh.c = context.WithValue(context.Background(), keyLog, l)
		} else {
			old = bh.c.Value(keyLog).(log.Logger)
		}

		prev = OptLogger(old)
		bh.c = context.WithValue(bh.c, keyLog, l.WithLocus("host"))

		return
	}
}

// OptTransport sets the net.Transport.
func OptTransport(t net.TransportFactory) Option {
	return func(bh *basicHost) (prev Option) {
		prev = OptTransport(bh.t)
		bh.t = t
		return
	}
}
