package net

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

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

// WriteTo a specified writer in big-endian format.
func (p Path) WriteTo(w io.Writer) (n int64, err error) {
	b := new(bytes.Buffer)
	binary.Write(b, binary.BigEndian, p.lenHdr())
	b.WriteString(p.String())
	return io.Copy(w, b)
}

// ReadFrom a specified reader and construct a Path.
func (p *Path) ReadFrom(r io.Reader) (n int64, err error) {
	var hdr uint16
	b := new(bytes.Buffer)

	if err = binary.Read(r, binary.BigEndian, &hdr); err != nil {
		err = errors.Wrap(err, "read len")
	} else if n, err = io.Copy(b, io.LimitReader(r, int64(hdr))); err != nil {
		err = errors.Wrap(err, "read path")
	} else {
		n += 2
		*p = Path(b.String())
	}

	return
}
