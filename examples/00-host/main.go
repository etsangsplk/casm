package main

import (
	"context"
	"fmt"
	"time"

	"github.com/lthibault/casm/pkg/host"
	"github.com/lthibault/casm/pkg/net"
	log "github.com/lthibault/log/pkg"
)

var c = context.Background()
var timeout = time.Second * 60

func main() {
	log := log.New(log.OptLevel(log.DebugLevel))

	h0 := host.New(host.OptLogger(log), host.OptListenAddr("/h0"))
	h0.Stream().Register("/echo", net.HandlerFunc(func(s *net.Stream) {
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

	h1 := host.New(host.OptLogger(log), host.OptListenAddr("/h1"))

	if err := h0.Start(c); err != nil {
		log.Fatal(err)
	}

	if err := h1.Start(c); err != nil {
		log.Fatal(err)
	}

	connCtx, cancel := context.WithTimeout(c, timeout)
	defer cancel()

	// Connect the hosts to each other
	if err := h0.Network().Connect(connCtx, h1); err != nil {
		log.Fatal(err)
	}

	s, err := h1.Stream().Open(h0, "/echo")
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

	fmt.Println(string(b))
}
