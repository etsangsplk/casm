package graph

import casm "github.com/lthibault/casm/pkg"

// compile-time type constraints
var _ Broadcaster = &broadcast{}

// Broadcaster handles message broadcast/receipt
type Broadcaster interface {
	Send([]byte) error
	Recv() ([]byte, error)
	Publish()
	Subscribe()
}

type broadcast struct {
	f messageFactory
}

func newBroadcaster(id casm.IDer) *broadcast {
	return &broadcast{f: newMsgFactory(id.ID())}
}

func (b broadcast) Send(msg []byte) error { return b.sendMsg(b.f(msg)) }

func (b broadcast) sendMsg(m *message) error {
	panic("sendMsg NOT IMPLEMENTED")
}

func (b broadcast) Recv() (msg []byte, err error) {
	var m *message
	defer m.Free()

	if m, err = b.recvMsg(); err == nil {
		msg = m.Body()
	}

	return
}

func (b broadcast) recvMsg() (*message, error) {
	panic("recvMsg NOT IMPLEMENTED")
}

func (b broadcast) Publish() {
	panic("Publish NOT IMPLEMENTED")
}

func (b broadcast) Subscribe() {
	panic("Subscribe NOT IMPLEMENTED")
}
