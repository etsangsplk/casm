package host

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"sync"
	"time"

	radix "github.com/armon/go-radix"
	net "github.com/lthibault/casm/pkg/net"
	log "github.com/lthibault/log/pkg"
	"github.com/pkg/errors"
)

// Stream is a full-duplex stream on top of a single physical connection to
// a remote host
type Stream interface {
	Path() string
	Context() context.Context
	StreamID() uint32
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close() error
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// Handler responds to an incoming stream connection
type Handler interface {
	Serve(Stream)
}

// HandlerFunc is an adapter to allow the use of ordinary functions as stream
// handlers.  If f is a function with the appropriate signature, HandlerFunc(f)
// is a Handler that calls f.
type HandlerFunc func(Stream)

// Serve satisfies Handler.  It calls h.
func (h HandlerFunc) Serve(s Stream) { h(s) }

// path to a stream
type path string

func (p path) lenHdr() uint16 { return uint16(len(p)) }

func (p path) String() string { return string(p) }

// SendTo a specified writer in big-endian format.
func (p path) SendTo(w io.Writer) (err error) {
	b := new(bytes.Buffer)
	binary.Write(b, binary.BigEndian, p.lenHdr())
	b.WriteString(p.String())
	_, err = io.Copy(w, b)
	return
}

// RecvFrom a specified reader and construct a path.
func (p *path) RecvFrom(r io.Reader) (err error) {
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
		*p = path(b.String())
	}

	return
}

type streamMux struct {
	lock sync.RWMutex
	log  log.Logger
	r    *radix.Tree
}

func newStreamMux(l log.Logger) *streamMux {
	return &streamMux{log: l, r: radix.New()}
}

func (m *streamMux) Register(path string, h Handler) {
	m.lock.Lock()
	m.log.WithField("path", path).Debug("registered handler")
	m.r.Insert(path, h)
	m.lock.Unlock()
}

func (m *streamMux) Unregister(path string) {
	m.lock.Lock()
	if _, ok := m.r.Delete(path); ok {
		m.log.WithField("path", path).Debug("unregistered")
	}
	m.lock.Unlock()
}

func (m *streamMux) Serve(s Stream) {
	m.lock.RLock()
	v, ok := m.r.Get(s.Path())
	m.lock.RUnlock()

	if ok {
		v.(Handler).Serve(s)
	}
}

type stream struct {
	path string
	*net.Stream
}

func (s stream) Path() string { return s.path }
