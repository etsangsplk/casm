package graph

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/lthibault/casm/api/graph"
	casm "github.com/lthibault/casm/pkg"
	capnp "zombiezen.com/go/capnproto2"
)

type messagePool sync.Pool

func (p *messagePool) Get() (m *message) {
	m = (*sync.Pool)(unsafe.Pointer(p)).Get().(*message)
	m.ctr = 1
	return
}

func (p *messagePool) Put(m *message) { (*sync.Pool)(unsafe.Pointer(p)).Put(m) }

var msgPool = messagePool(sync.Pool{New: func() interface{} {
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
}})

type message struct {
	cm  *capnp.Message
	m   graph.Message
	ctr uint32
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
		msgPool.Put(m)
	}
}

type messageFactory func([]byte) *message

func newMsgFactory(pid casm.PeerID) func([]byte) *message {
	var seq uint64
	return func(b []byte) (msg *message) {
		msg = msgPool.Get()
		msg.m.SetId(uint64(pid))
		msg.m.SetSeq(atomic.AddUint64(&seq, 1))
		msg.m.SetBody(b)
		return
	}
}
