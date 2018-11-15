package casm

// Handler responds to an incoming stream connection
type Handler interface {
	ServeStream(Stream)
}

// HandlerFunc is an adapter to allow the use of ordinary functions as stream
// handlers.  If f is a function with the appropriate signature, HandlerFunc(f)
// is a Handler that calls f.
type HandlerFunc func(Stream)

// ServeStream satisfies Handler.  It calls h.
func (h HandlerFunc) ServeStream(s Stream) { h(s) }
