package graph

// Broadcaster handles message broadcast/receipt
type Broadcaster interface {
	Send()
	Recv()
	Publish()
	Subscribe()
}
