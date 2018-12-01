package host

import (
	net "github.com/lthibault/casm/pkg/net"
	log "github.com/lthibault/log/pkg"
	"github.com/satori/uuid"
)

// Option represents a setting
type Option func(*cfg) Option

// OptListenAddr sets the listen address
func OptListenAddr(addr string) Option {
	return func(c *cfg) (prev Option) {
		if c.Addr != nil {
			prev = OptListenAddr(c.Addr.String())
		}
		c.Addr = net.NewAddr(0, "inproc", addr)
		return
	}
}

// OptLogger sets the logger
func OptLogger(log log.Logger) Option {
	return func(c *cfg) (prev Option) {
		prev = OptLogger(c.Logger)
		c.Logger = log
		return
	}
}

type cfg struct {
	net.Addr
	log.Logger
}

func mkHost(c cfg) *basicHost {
	id := net.New()
	if c.Addr == nil {
		net.NewAddr(id, "inproc", "/"+uuid.NewV4().String())
	}

	bh := new(basicHost)
	bh.a = net.NewAddr(id, c.Addr.Network(), c.Addr.String())
	bh.t = net.NewTransport(id, c.Addr)
	bh.log = mkLogger(c, id)
	bh.mux = newStreamMux(bh.log.WithLocus("mux"))
	bh.peers = newPeerStore()

	return bh
}

func mkLogger(c cfg, id net.PeerID) log.Logger {
	l := c.Logger
	if l == nil {
		l = log.New()
	}
	return l.WithFields(log.F{"id": id, "locus": "host"})
}
