package main

import (
	"context"
	"log"
	"time"

	casm "github.com/lthibault/casm/pkg"
)

func main() {
	h0, err := casm.New(context.Background(), casm.OptListenAddrStrings("/ip4/0.0.0.0/tcp/9021"))
	if err != nil {
		panic(err)
	}

	h1, err := casm.New(context.Background(), casm.OptListenAddrStrings("/ip4/0.0.0.0/tcp/9022"))
	if err != nil {
		panic(err)
	}

	h0.RegisterStreamHandler("/echo", casm.HandlerFunc(func(s casm.Stream) {
		defer s.Close() // You MUST call Close to avoid memory leaks

		b := make([]byte, 11)
		for {
			select {
			case <-s.Context().Done():
				return
			default:
				n, err := s.Read(b)
				if err != nil {
					panic(err)
				}

				if _, err = s.Write(b[:n]); err != nil {
					panic(err)
				}
			}
		}
	}))

	// Connect the hosts to each other
	if err = h0.Connect(context.Background(), h1); err != nil {
		panic(err)
	}

	// Open a stream
	c, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	s, err := h1.OpenStream(c, h0, "/echo")
	if err != nil {
		panic(err)
	}

	if _, err = s.Write([]byte("hello world")); err != nil {
		panic(err)
	}

	b := make([]byte, 11)
	if _, err = s.Read(b); err != nil {
		panic(err)
	}

	log.Fatal(string(b))
}
