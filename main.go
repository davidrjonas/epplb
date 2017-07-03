package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
)

var (
	listen   = flag.String("listen", ":10700", "target")
	upstream = flag.String("upstream", "epp-ote.verisign-grs.com:700", "Upstream to which we should proxy")
	certFile = flag.String("cert", "crt.pem", "A PEM eoncoded certificate file.")
	keyFile  = flag.String("key", "key.pem", "A PEM encoded private key file.")
	caFile   = flag.String("ca", "ca.pem", "A PEM eoncoded CA's certificate file.")
)

func main() {
	flag.Parse()

	server, err := net.Listen("tcp", *listen)

	if err != nil {
		log.Fatalf("Failed to listen; address=%s, err=%v", *listen, err)
	}

	p, _ := NewProxy(server, makeTlsClientFactory(*upstream, *certFile, *keyFile, *caFile))

	stop := make(chan bool, 1)
	go p.Listen(stop)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	// Block until a signal is received.
	<-sigs

	fmt.Println("Closing listener and waiting for clients to finish")
	stop <- true
}
