package main

import (
	"log"
	"net"
	"sync"
	"sync/atomic"

	"gopkg.in/fatih/pool.v2"
)

type Proxy struct {
	pool    pool.Pool
	server  net.Listener
	counter uint32
}

type accepted struct {
	c   net.Conn
	err error
}

// NewProxy creates a new proxy.
func NewProxy(server net.Listener, upstreamFactory func() (net.Conn, error)) (*Proxy, error) {
	upstreams, err := pool.NewChannelPool(1, 8, upstreamFactory)
	if err != nil {
		return nil, err
	}

	p := Proxy{
		server: server,
		pool:   upstreams,
	}

	return &p, nil
}

func (p *Proxy) Listen(stop <-chan bool) {
	var wg sync.WaitGroup

	accept := make(chan accepted, 1)

	go func() {
		for {
			conn, err := p.server.Accept()
			accept <- accepted{conn, err}
		}
	}()

	for {
		select {
		case a := <-accept:
			atomic.AddUint32(&p.counter, 1)
			if a.err != nil {
				log.Printf("Failed when accepting client; %v", a.err)
				break
			}

			upstream, err := p.pool.Get()

			if err != nil {
				log.Printf("Failed to get a connection from the pool; %v", err)
				a.c.Close()
				break
			}

			wg.Add(1)

			pair := ProxyPair{upstream: upstream.(*pool.PoolConn), client: a.c.(*net.TCPConn)}
			go pair.talk(wg.Done)
		case <-stop:
			break
		}
	}

	// Close the listener but let the clients drain
	p.server.Close()
	wg.Wait()
}
