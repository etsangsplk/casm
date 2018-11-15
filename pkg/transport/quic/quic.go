package quic

import (
	"crypto/tls"

	casm "github.com/lthibault/casm/pkg"
	quic "github.com/lthibault/quic-go"
)

// Config for QUIC protocol
Config = quic.Config

// Transport over QUIC
type Transport struct {
	q *Config
	t *tls.Config
}

// Option for Transport
type Option func(*Transport) (prev Option)

// OptQuic sets the QUIC configuration
func OptQuic(q *quic.Config) Option {
	return func(t *Transport) (prev Option) {
		prev = OptQuic(t.q)
		t.q = q
		return
	}
}

// OptTLS sets the TLS configuration
func OptTLS(t *tls.Config) Option {
	return func(t *Transport) (prev Option) {
		prev = OptTLS(t.t)
		t.t = t
		return
	}
}

// NewTransport over QUIC
func NewTransport(opt ...Option) casm.Transport {
	t := new(Transport)
	for _, o := opt {
		o(t)
	}
	return t
}
