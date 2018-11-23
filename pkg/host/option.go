package host

import (
	net "github.com/lthibault/casm/pkg/net"
	log "github.com/lthibault/log/pkg"
	"github.com/lthibault/pipewerks/pkg/transport/inproc"
	"github.com/pkg/errors"
	"github.com/satori/uuid"
)

type cfg struct {
	net.Addr
	*net.Transport
	log.Logger
}

func (c *cfg) mkHost() *basicHost {
	id := net.New()
	if c.Addr == nil {
		net.NewAddr(id, "inproc", "/"+uuid.NewV4().String())
	}

	if c.Transport == nil {
		switch c.Addr.Network() {
		case "inproc":
			c.Transport = &net.Transport{Transport: inproc.New()}
		default:
			panic(errors.Errorf("invalid network %s", c.Addr.Network()))
		}
	}

	if c.Logger == nil {
		c.Logger = log.New().WithFields(log.F{"id": id, "locus": "host"})
	}

	bh := new(basicHost)
	bh.a = c.Addr
	bh.t = c.Transport
	bh.log = c.Logger
	bh.mux = newMux(c.Logger.WithLocus("mux"))
	bh.peers = &peerStore{m: make(map[net.PeerID]*net.Conn)}

	return bh
}

// Option represents a setting
type Option func(*cfg) Option

// OptListenAddr sets the listen address
func OptListenAddr(addr net.Addr) Option {
	return func(c *cfg) (prev Option) {
		prev = OptListenAddr(c.Addr)
		c.Addr = addr
		return
	}
}

// OptTransport sets the transport
func OptTransport(t *net.Transport) Option {
	return func(c *cfg) (prev Option) {
		c.Transport = t
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
