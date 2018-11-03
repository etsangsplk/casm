package main

import (
	"context"
	"log"
	"time"

	casm "github.com/lthibault/casm/pkg"
)

func main() {
	h0, err := casm.New(context.Background(), casm.OptListenAddrStrings("/ip4/127.0.0.1/tcp/9021"))
	if err != nil {
		log.Fatal(err)
	}

	h1, err := casm.New(context.Background(), casm.OptListenAddrStrings("/ip4/127.0.0.1/tcp/9022"))
	if err != nil {
		log.Fatal(err)
	}

	h0.Stream().Register("/echo", casm.HandlerFunc(func(s casm.Stream) {
		defer s.Close() // You MUST call Close to avoid memory leaks

		b := make([]byte, 11)
		for {
			select {
			case <-s.Context().Done():
				return
			default:
				n, err := s.Read(b)
				if err != nil {
					log.Fatal(err)
				}

				if _, err = s.Write(b[:n]); err != nil {
					log.Fatal(err)
				}
			}
		}
	}))

	// Connect the hosts to each other
	if err = h0.Network().Connect(context.Background(), h1); err != nil {
		log.Fatal(err)
	}

	// Open a stream
	c, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	s, err := h1.Stream().Open(c, h0, "/echo")
	if err != nil {
		log.Fatal(err)
	}

	if _, err = s.Write([]byte("hello world")); err != nil {
		log.Fatal(err)
	}

	b := make([]byte, 11)
	if _, err = s.Read(b); err != nil {
		log.Fatal(err)
	}

	log.Fatal(string(b))
}
