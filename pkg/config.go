package casm

import (
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/config"

	ma "github.com/multiformats/go-multiaddr"
)

// Option represents a setting
type Option interface {
	Opt()
}

/***************************************
Adapters for libp2p Host options
****************************************/

// p2pOpt allows for the annotation of libp2p Options so that they can be
// recognized by New() and passed to libp2p.New().
type p2pOpt func(*config.Config) error

func (p2pOpt) Opt() {}

type p2pOptions []libp2p.Option

func defaultP2pOpts() p2pOptions {
	panic("defaultLibp2pOpts NOT IMPLEMENTED")
}

func (h p2pOptions) Load(opt []Option) {
	for _, o := range opt {
		if op, ok := o.(p2pOpt); ok {
			h = append(h, libp2p.Option(op))
		}
	}
}

// ListenAddrStrings configures the CASM host to listen on the given multiaddrs
func ListenAddrStrings(s ...string) Option {
	return p2pOpt(libp2p.ListenAddrStrings(s...))
}

// ListenAddrs configures the CASM host to listen on the given multiaddrs
func ListenAddrs(addrs ...ma.Multiaddr) Option {
	return p2pOpt(libp2p.ListenAddrs(addrs...))
}

/************************
	CASM Host options
*************************/

type hostOpt func(*Host) error

func (hostOpt) Opt()                    {}
func (opt hostOpt) Apply(h *Host) error { return opt(h) }

type hostOptions []hostOpt

func defaultHostOpts() hostOptions {
	panic("defaultHostOpts NOT IMPLEMENTED")
}

func (h hostOptions) Load(opt []Option) {
	for _, o := range opt {
		if op, ok := o.(hostOpt); ok {
			h = append(h, op)
		}
	}
}

// Cardinality configures the number of
func Cardinality(k uint8) Option {
	return hostOpt(func(h *Host) error {
		h.k = k
		return nil
	})
}
