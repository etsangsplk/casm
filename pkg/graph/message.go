package graph

import (
	"sync"
	"sync/atomic"

	"github.com/lthibault/casm/api/graph"
	casm "github.com/lthibault/casm/pkg"
	capnp "zombiezen.com/go/capnproto2"
)

var ( // compile-time type constraints
	_ Message = new(message)
	// _ SubMessage = ...
)

// Message for broadcast over the graph
type Message interface {
	ID() casm.PeerID
	Sequence() uint64
	Header() []byte
	Body() []byte
	Ref()
	Free()
}

// SubMessage is a message for a subscription
type SubMessage interface {
	Message
	Topic() []byte
}

var msgPool = sync.Pool{New: func() interface{} {
	var err error
	msg := new(message)

	seg := new(capnp.Segment)
	if msg.cm, seg, err = capnp.NewMessage(capnp.SingleSegment(nil)); err != nil {
		panic(err)
	}

	if msg.m, err = graph.NewRootMessage(seg); err != nil {
		panic(err)
	}

	return msg
}}

type message struct {
	cm  *capnp.Message
	m   graph.Message
	ctr uint32
	p   *sync.Pool
}

// ID of the sender
func (m message) ID() casm.PeerID { return casm.PeerID(m.m.Id()) }

// Sequence number of the message
func (m message) Sequence() uint64 { return uint64(m.m.Seq()) }

// Header uniquely idenitifies a message
func (m message) Header() []byte {
	panic("Header NOT IMPLEMENTED")
}

// Body of the message
func (m message) Body() []byte {
	b, err := m.m.Body()
	if err != nil {
		panic(err)
	}
	return b
}

// Ref increases the reference count for the message
func (m *message) Ref() { atomic.AddUint32(&m.ctr, 1) }

// Free a reference count for the message.  When the refcount hits zero, the
// message will be returned to the sync.Pool.
func (m *message) Free() {
	if atomic.AddUint32(&m.ctr, ^uint32(0)) == 0 {
		m.p.Put(m)
	}
}

func messageFactory(pid casm.PeerID) func([]byte) *message {
	var seq uint64
	return func(b []byte) (msg *message) {
		msg = msgPool.Get().(*message)
		msg.p = &msgPool
		msg.ctr = 1
		msg.m.SetId(uint64(pid))
		msg.m.SetSeq(atomic.AddUint64(&seq, 1))
		msg.m.SetBody(b)
		return
	}
}
