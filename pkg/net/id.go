package net

import (
	"fmt"
	"math/rand"
	"time"
	"unsafe"

	"github.com/aybabtme/rgbterm"
)

func init() { rand.Seed(time.Now().UTC().UnixNano()) }

// PeerID is a unique identifier for a Node
type PeerID uint64

// New instance
func New() PeerID { return PeerID(rand.Uint64()) }

func (id PeerID) String() string { return fmt.Sprintf("%016x", uint64(id)) }

// ID satisfies the IDer interface
func (id PeerID) ID() PeerID { return id }

// Color produces bytes which, when passed to a terminal, produced a colorized
// rendering of the PeerID.  Color is decoded from the first 3 bytes of the
// PeerID.
func (id PeerID) Color() string {
	arr := (*[8]uint8)(unsafe.Pointer(&id))
	r := arr[0]
	g := arr[1]
	b := arr[2]
	return rgbterm.FgString(id.String(), r, g, b)
}
