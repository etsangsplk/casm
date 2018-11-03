package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/lthibault/casm/pkg"
	"github.com/lthibault/casm/pkg/graph"
	"github.com/pkg/errors"
)

const k = 3

var (
	c     = context.Background()
	addrs = []string{
		"/ip4/127.0.0.1/tcp/9021",
		"/ip4/127.0.0.1/tcp/9022",
		"/ip4/127.0.0.1/tcp/9023",
		"/ip4/127.0.0.1/tcp/9024",
		"/ip4/127.0.0.1/tcp/9025",
	}

	// We'll store the most recent message here, so that we can check that
	// the broadcast was received by all.
	lock sync.Mutex
	res  = make(map[int]string, len(addrs))
)

func newVertex(listen string) (graph.Vertex, error) {
	var h casm.Host
	if h, err = casm.New(c, casm.OptListenAddrStrings(l)); err != nil {
		err = errors.Wrap(err, "create host")
		break
	}

	if v, err = graph.New(h, graph.OptCardinality(k)); err != nil {
		err = errors.Wrap(err, "create vertex")
		break
	}
}

func buildGraph(listen ...string) (vtxs []graph.Vertex, err error) {
	vtxs = make([]graph.Vertex, len(listen))

	for i, l := range listen {
		if vtxs[i], err = newVertex(l); err != nil {
			break
		}

		if i > 0 {
			if err = vtx[i].Edge().Lease(c), vtx[i-1]; err != nil {
				err = errors.Wrap(err, "lease")
				break
			}
		}
	}

	return
}

func broadcast(wg *sync.WaitGroup, id int, v graph.Vertex) {
	defer wg.Done()

	// we'll receive in the background.  This is sloppy since the goroutine is
	// never explicitly stopped, but this will suffice.
	go func() {
		for {
			b, err := v.Message().Recv()
			if err != nil {
				panic(err)
			}

			lock.Lock()
			res[id] = string(b)
			res.Unlock()
		}
	}()

	for _, msg := range []string{"alpha", "bravo", "charlie", "delta", "echo"} {
		b := []byte(fmt.Sprintf("%d:%s", id, msg))
		if err := v.Message().Send(b); err != nil {
			panic(err)
		}
	}
}

func main() {
	vtxs, err := buildGraph(addrs...)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(len(addrs))
	for i, v := range vtxs {
		go broadcast(&wg, i, v)
	}
	wg.Wait()

	// let's wait another couple of ms to ensure all messages were actually
	// *received*.  Sloppy, but saves us the trouble of complex synchronization.
	<-time.After(time.Millisecond * 10)

	// Let's *NOT* lock the result map.  If someone writes while we're reading,
	// then we didn't wait long enough, and we should panic.  Again, this is
	// sloppy, but expedient.
	for _, msg := range res {
		if msg != "echo" {
			log.Fatalf("unexpected message %s", msg)
		}
	}
}
