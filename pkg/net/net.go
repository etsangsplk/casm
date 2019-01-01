// Package net provides a standard interface for network IO throughout CASM.
package net

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

func init() { rand.Seed(time.Now().UTC().UnixNano()) }

// PeerID is a unique identifier for a Node
type PeerID uint64

// New instance
func New() PeerID { return PeerID(rand.Uint64()) }

func (id PeerID) String() string { return fmt.Sprintf("%016x", uint64(id)) }

// ID satisfies the IDer interface
func (id PeerID) ID() PeerID { return id }

// Handler responds to an incoming stream connection
type Handler interface {
	Serve(*Stream)
}

// HandlerFunc is an adapter to allow the use of ordinary functions as stream
// handlers.  If f is a function with the appropriate signature, HandlerFunc(f)
// is a Handler that calls f.
type HandlerFunc func(*Stream)

// Serve satisfies Handler.  It calls h.
func (h HandlerFunc) Serve(s *Stream) { h(s) }

// Path to a stream
type Path string

func (p Path) lenHdr() uint16 { return uint16(len(p)) }

func (p Path) String() string { return string(p) }

// SendTo a specified writer in big-endian format.
func (p Path) SendTo(w io.Writer) (err error) {
	b := new(bytes.Buffer)
	binary.Write(b, binary.BigEndian, p.lenHdr())
	b.WriteString(p.String())
	_, err = io.Copy(w, b)
	return
}

// RecvFrom a specified reader and construct a Path.
func (p *Path) RecvFrom(r io.Reader) (err error) {
	var n int64
	var hdr uint16
	b := new(bytes.Buffer)

	if err = binary.Read(r, binary.BigEndian, &hdr); err != nil {
		err = errors.Wrap(err, "read len")
	} else if n, err = io.Copy(b, io.LimitReader(r, int64(hdr))); err != nil {
		err = errors.Wrap(err, "read path")
	} else if uint16(n) != hdr { // EOF not handled by io.Copy
		err = errors.Wrap(io.EOF, "read path")
	} else {
		*p = Path(b.String())
	}

	return
}
