package host

import (
	"fmt"
	"net/url"

	net "github.com/lthibault/casm/pkg/net"
	log "github.com/lthibault/log/pkg"
	"github.com/satori/uuid"
)

// Option represents a setting
type Option func(*cfg) Option

// OptListenAddr sets the listen address
func OptListenAddr(addr string) Option {
	u, err := url.Parse(addr)
	if err != nil {
		panic(err)
	}

	netloc := fmt.Sprintf("%s%s", u.Host, u.EscapedPath())

	return func(c *cfg) (prev Option) {
		switch {
		case c.network == "":
			c.network = "inproc"
			fallthrough
		case c.addr == "":
			c.addr = fmt.Sprintf("/%s", uuid.NewV4())
		}

		prev = OptListenAddr(fmt.Sprintf("%s://%s", c.network, c.addr))

		c.network = u.Scheme
		c.addr = netloc

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
	network, addr string
	log.Logger
}

func mkHost(c cfg) *basicHost {
	laddr := c.ListenAddr(net.New())

	bh := new(basicHost)
	bh.a = laddr
	bh.t = net.NewTransport(laddr)
	bh.log = mkLogger(c, laddr.ID())
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

func (c cfg) ListenAddr(id net.PeerID) net.Addr {
	return net.NewAddr(id, c.network, c.addr)
}
