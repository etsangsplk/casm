package net

import (
	bin "encoding/binary"
	"io"
)

// binary is a big-endian binary codec.
// it exists as a convenience to avoid having to specify endianness during IO.
var binary = binaryCodec{}

type binaryCodec struct{}

func (binaryCodec) Write(w io.Writer, v interface{}) error {
	return bin.Write(w, bin.BigEndian, v)
}

func (binaryCodec) Read(r io.Reader, v interface{}) error {
	return bin.Read(r, bin.BigEndian, v)
}
