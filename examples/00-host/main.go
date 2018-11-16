package main

import (
	"context"
	"time"

	"github.com/lthibault/casm/pkg/host"
	clog "github.com/lthibault/casm/pkg/log"
	"github.com/lthibault/casm/pkg/net"
	log "github.com/sirupsen/logrus"
)

var c = context.Background()
var timeout = time.Second * 60

var l clog.Logger

func init() {
	lgrs := log.New()
	lgrs.SetLevel(log.DebugLevel)
	l = clog.WrapLogrus(lgrs)
}

func main() {
	h0 := host.New(host.OptLogger(l), host.OptListenAddr("localhost:9021"))
	if err := h0.ListenAndServe(c); err != nil {
		log.Fatal(err)
	}

	h1 := host.New(host.OptListenAddr("localhost:9022"))
	if err := h1.ListenAndServe(c); err != nil {
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

	connCtx, cancel := context.WithTimeout(c, timeout)
	defer cancel()

	// Connect the hosts to each other
	if err := h0.Network().Connect(connCtx, h1); err != nil {
		log.Fatal(err)
	}

	streamCtx, cancel := context.WithTimeout(c, timeout)

	// Open a stream
	c, cancel := context.WithTimeout(streamCtx, time.Second)
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
