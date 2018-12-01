package host

import (
	"context"
	"sync"

	radix "github.com/armon/go-radix"
	net "github.com/lthibault/casm/pkg/net"
	log "github.com/lthibault/log/pkg"
)

type streamMux struct {
	lock sync.RWMutex
	log  log.Logger
	r    *radix.Tree
}

func newStreamMux(l log.Logger) *streamMux {
	return &streamMux{log: l, r: radix.New()}
}

func (m *streamMux) Register(c context.Context, path string, h net.Handler) {
	m.lock.Lock()
	m.log.WithField("path", path).Debug("registered")
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

func (m *streamMux) Serve(path string, s *net.Stream) {
	m.lock.RLock()
	if v, ok := m.r.Get(path); ok {
		go v.(net.Handler).Serve(s)
	}
	m.lock.RUnlock()
}
