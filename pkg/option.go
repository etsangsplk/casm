package casm

import (
	"github.com/libp2p/go-libp2p"
	net "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p/config"

	ma "github.com/multiformats/go-multiaddr"
)

// NetHook is a set of callbacks that are invoked according to network changes
type NetHook = net.Notifiee

// Option represents a setting
type Option interface {
	opt()
}

// Applicator can be applied to generic Hosts
type Applicator interface {
	Option
	Apply(Host) error
}

/***************************************
Adapters for libp2p Host options
****************************************/

// p2pOpt allows for the annotation of libp2p Options so that they can be
// recognized by New() and passed to libp2p.New().
type p2pOpt func(*config.Config) error

func (p2pOpt) opt() {}

type p2pOptions []libp2p.Option

func defaultP2pOpts() p2pOptions {
	return []libp2p.Option{}
}

func (h *p2pOptions) Load(opt []Option) {
	for _, o := range opt {
		if op, ok := o.(p2pOpt); ok {
			*h = append(*h, libp2p.Option(op))
		}
	}
}

// OptListenAddrStrings configures the CASM host to listen on the given multiaddrs
func OptListenAddrStrings(s ...string) Option {
	return p2pOpt(libp2p.ListenAddrStrings(s...))
}

// OptListenAddrs configures the CASM host to listen on the given multiaddrs
func OptListenAddrs(addrs ...ma.Multiaddr) Option {
	return p2pOpt(libp2p.ListenAddrs(addrs...))
}

/************************
	CASM Host options
*************************/

type hostOpt func(*basicHost) error

func (hostOpt) opt()                   {}
func (opt hostOpt) Apply(h Host) error { return opt(h.(*basicHost)) }

type hostOptions []hostOpt

func defaultHostOpts() hostOptions {
	return []hostOpt{}
}

func (h *hostOptions) Load(opt []Option) {
	for _, o := range opt {
		if op, ok := o.(hostOpt); ok {
			*h = append(*h, op)
		}
	}
}

// OptNetHook sets a NetHook on the host
func OptNetHook(h NetHook) Option {
	return hostOpt(func(b *basicHost) error {
		b.h.Network().Notify(h)
		return nil
	})
}
