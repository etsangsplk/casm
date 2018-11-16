package host

import (
	"crypto/tls"

	log "github.com/lthibault/casm/pkg/log"
	net "github.com/lthibault/casm/pkg/net"
	quic "github.com/lthibault/casm/pkg/transport/quic"
	"github.com/pkg/errors"
)

type cfg struct {
	Addr              string
	CertPath, KeyPath string
	Insecure          bool

	id net.PeerID
	t  net.Transport
	l  log.Logger
	qc *quic.Config
	tc *tls.Config
}

func (c *cfg) mkHost() *basicHost {
	if c.t == nil {
		if c.CertPath == "" && c.KeyPath == "" {
			if !c.Insecure {
				panic(errors.New("TLS required"))
			}
			c.tc = generateTLSConfig()
		} else {
			panic("LoadTLS NOT IMPLEMENTED")
		}

		if c.qc == nil {
			c.qc = defaultQUIC()
		}

		c.t = quic.New(c.id, quic.OptQuic(c.qc), quic.OptTLS(c.tc))
	}

	bh := new(basicHost)
	bh.addr = c.Addr
	bh.id = c.id
	bh.t = c.t
	bh.log = c.l.WithField("id", c.id).WithLocus("host")
	bh.mux = newMux()
	bh.peers = &peerStore{m: make(map[net.PeerID]net.Conn)}

	return bh
}

// Option represents a setting
type Option func(*cfg) Option

// OptListenAddr sets the listen address
func OptListenAddr(addr string) Option {
	return func(c *cfg) (prev Option) {
		c.Addr = addr
		return
	}
}

// OptTransport sets the transport
func OptTransport(t net.Transport) Option {
	return func(c *cfg) (prev Option) {
		c.t = t
		return
	}
}

// OptLogger sets the logger
func OptLogger(log log.Logger) Option {
	return func(c *cfg) (prev Option) {
		prev = OptLogger(c.l)
		c.l = log
		return
	}
}

func optSetID() Option {
	return func(c *cfg) (prev Option) {
		prev = optSetID()
		c.id = net.New()
		return
	}
}

func defaultQUIC() *quic.Config {
	return &quic.Config{
		KeepAlive: true,
	}
}

func maybeMkQUIC() Option {
	return func(c *cfg) (prev Option) {
		prev = OptTransport(c.t)
		if c.t == nil {
			tc := generateTLSConfig()
			qc := defaultQUIC()

			c.t = quic.New(c.id, quic.OptQuic(qc), quic.OptTLS(tc))
		}
		return
	}
}

func defaultOpts(overrides ...Option) []Option {
	opt := []Option{
		OptLogger(log.NoOp()),
		optSetID(),
		OptListenAddr("localhost:1987"),
	}

	overrides = append(overrides, maybeMkQUIC())
	return append(opt, overrides...)
}
