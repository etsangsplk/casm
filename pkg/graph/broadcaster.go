package graph

import (
	"context"
	"sync"
	"unsafe"

	"github.com/lthibault/portal/protocol/bus"

	"github.com/lthibault/portal"

	casm "github.com/lthibault/casm/pkg"
)

// compile-time type constraints
var _ Broadcaster = &broadcast{}

const bufSize = 8

var fp = framePool(sync.Pool{New: func() interface{} { return new(edgeFrame) }})

// Broadcaster handles message broadcast/receipt
type Broadcaster interface {
	Send([]byte) error
	Recv() ([]byte, error)
	Publish()
	Subscribe()
}

type framePool sync.Pool

func (p *framePool) Get(e Edge) (f *edgeFrame) {
	f = (*sync.Pool)(unsafe.Pointer(p)).Get().(*edgeFrame)
	f.Edge = e
	f.c, f.cancel = context.WithCancel(e.Context())
	return
}

func (p *framePool) Put(f *edgeFrame) {
	f.cancel()
	f.Edge = nil
	f.c = nil
	f.cancel = nil
	(*sync.Pool)(unsafe.Pointer(p)).Put(f)
}

type edgeFrame struct {
	Edge
	c      context.Context
	cancel func()
}

func (f edgeFrame) Context() context.Context { return f.c }
func (f *edgeFrame) Free()                   { fp.Put(f) }

type edgeSet map[casm.PeerID]*edgeFrame

func (es edgeSet) Add(f *edgeFrame) { es[f.RemotePeer()] = f }
func (es edgeSet) Get(id casm.IDer) (f *edgeFrame, ok bool) {
	f, ok = es[id.ID()]
	return
}

type broadcast struct {
	sync.RWMutex
	es edgeSet
	f  messageFactory
	p  portal.Portal
}

func newBroadcaster(id casm.IDer) *broadcast {
	return &broadcast{
		f:  newMsgFactory(id.ID()),
		es: make(map[casm.PeerID]*edgeFrame),
		p:  portal.New(bus.New(), portal.OptAsync(bufSize)),
	}
}

func (b *broadcast) Send(msg []byte) error { return b.sendMsg(b.f(msg)) }

func (b *broadcast) sendMsg(m *message) error {
	panic("sendMsg NOT IMPLEMENTED")
}

func (b *broadcast) Recv() (msg []byte, err error) {
	var m *message
	defer m.Free()

	if m, err = b.recvMsg(); err == nil {
		msg = m.Body()
	}

	return
}

func (b *broadcast) recvMsg() (*message, error) {
	panic("recvMsg NOT IMPLEMENTED")
}

func (b *broadcast) Publish() {
	panic("Publish NOT IMPLEMENTED")
}

func (b *broadcast) Subscribe() {
	panic("Subscribe NOT IMPLEMENTED")
}

func (b *broadcast) AddEdge(e Edge) {
	b.Lock()
	defer b.Unlock()

	b.es.Add(fp.Get(e))

	//
	//  TODO:  implement *edge.  Ensure its contexts properly terminate.
	//
	panic("AddEdge NOT IMPLEMENTED")

	// go func() {
	// 	ch := b.p.Open()
	// 	defer ch.Close()

	// 	for range ctx.Tick(f.Context()) {
	// 		select {
	// 		case msg, ok := <-ch.Recv():
	// 			if !ok {
	// 				continue
	// 			}

	// 			if err := msg.(*message).WriteTo(e.Pipe()); err != nil {
	// 				// NOTE:  edge should detect non-temporary net.Errors and
	// 				// 		  close the pipe completely, which should terminate
	// 				// 		  the context. No manual intervention should be
	// 				//		  necessary.
	// 				// TODO:  double-check that the above was properly implemented
	// 				panic("TODO:  log the error")
	// 			}
	// 		}
	// 		panic("TODO:  receive messages from edges")
	// 	}
	// }()
}

func (b *broadcast) RemoveEdge(id casm.IDer) (e Edge, ok bool) {
	b.Lock()
	defer b.Unlock()

	var f *edgeFrame
	if f, ok = b.es.Get(id); ok {
		e = f.Edge
		f.Free()
	}
	return
}
