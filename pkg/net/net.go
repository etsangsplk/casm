package net

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
