package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/davidrjonas/epplb/rfc5734"

	pool "gopkg.in/fatih/pool.v2"
)

var (
	listen   = flag.String("listen", ":10700", "target")
	upstream = flag.String("upstream", "epp-ote.verisign-grs.com:700", "Upstream to which we should proxy")
	certFile = flag.String("cert", "crt.pem", "A PEM eoncoded certificate file.")
	keyFile  = flag.String("key", "key.pem", "A PEM encoded private key file.")
	caFile   = flag.String("ca", "ca.pem", "A PEM eoncoded CA's certificate file.")
	maxConns = flag.Int("max-conns", 1, "Maximum number of upstream connections to open")
)

func mustCreatePool(capacity int, upstreamHost, certFile, keyFile, caFile string) pool.Pool {
	upstreams, err := pool.NewChannelPool(1, capacity, NewTlsClientFactory(upstreamHost, certFile, keyFile, caFile))

	if err != nil {
		log.Fatalf("Failed to create pool; %v", err)
	}

	return upstreams
}

func mustListen(laddr string) net.Listener {
	server, err := net.Listen("tcp", laddr)

	if err != nil {
		log.Fatalf("Failed to listen; address=%s, err=%v", laddr, err)
	}

	return server
}

func NewEppServer(laddr string, maxConns int, upstreamHost, certFile, keyFile, caFile string) *rfc5734.Server {
	h := ProxyHandler{pool: mustCreatePool(maxConns, upstreamHost, certFile, keyFile, caFile), MaxRetries: 3}
	s := rfc5734.NewServer(mustListen(laddr))

	go s.Serve(h.Handle)

	return s
}

func main() {
	flag.Parse()

	s := NewEppServer(*listen, *maxConns, *upstream, *certFile, *keyFile, *caFile)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	<-sigs

	log.Println("Closing listener and waiting for clients to finish")
	s.Stop()
}
