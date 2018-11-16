package main

import (
	"context"
	"log"
	"time"

	"github.com/lthibault/casm/pkg/host"
	"github.com/lthibault/casm/pkg/net"
)

func main() {
	h0, err := host.New(context.Background(), host.OptListenAddr("localhost:9021"))
	if err != nil {
		log.Fatal(err)
	}

	h1, err := host.New(context.Background(), host.OptListenAddr("localhost:9022"))
	if err != nil {
		log.Fatal(err)
	}

	h0.Stream().Register("/echo", net.HandlerFunc(func(s net.Stream) {
		defer s.Close() // Users SHOULD close streams explicitly

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
