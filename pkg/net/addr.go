package net

import (
	"io"
	"net"

	"github.com/lunixbochs/struc"
)

// Addr of a Host
type Addr interface {
	ID() PeerID
	Addr() Addr
	// Proto indicates the transport protocol, e.g.:  "tcp", "quic", "utp", ...
	Proto() string
	net.Addr
}

type addr struct {
	PeerID
	proto, network, addr string
}

// NewAddr from an ID and an address stringer
func NewAddr(id PeerID, net, p, a string) Addr {
	return &addr{PeerID: id, network: net, proto: p, addr: a}
}

func (a addr) Addr() Addr      { return a }
func (a addr) Network() string { return a.network }
func (a addr) Proto() string   { return a.proto }
func (a addr) String() string  { return a.addr }

type wireAddr struct {
	PID      PeerID `struc:"uint64"`
	NetLen   int    `struc:"uint8,sizeof=NetStr"`
	NetStr   string
	ProtoLen int `struc:"uint8,sizeof=ProtoStr"`
	ProtoStr string
	AddrLen  int `struc:"uint8,sizeof=AddrStr"`
	AddrStr  string
}

func newWireAddr(a Addr) *wireAddr {
	return &wireAddr{
		PID:      a.ID(),
		NetLen:   len(a.Network()),
		NetStr:   a.Network(),
		ProtoLen: len(a.Proto()),
		ProtoStr: a.Proto(),
		AddrLen:  len(a.String()),
		AddrStr:  a.String(),
	}
}

func (a wireAddr) ID() PeerID      { return a.PID }
func (a wireAddr) Addr() Addr      { return a }
func (a wireAddr) Network() string { return a.NetStr }
func (a wireAddr) Proto() string   { return a.ProtoStr }
func (a wireAddr) String() string  { return a.AddrStr }

func (a *wireAddr) RecvFrom(r io.Reader) error { return struc.Unpack(r, a) }
func (a *wireAddr) SendTo(w io.Writer) error   { return struc.Pack(w, a) }
