package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
)

func NewTlsClientFactory(address, certFile, keyFile, caFile string) func() (net.Conn, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		log.Fatalf("Failed to load cert and key; certFile=%s, keyFile=%s, err=%v", certFile, keyFile, err)
	}

	caCert, err := ioutil.ReadFile(caFile)

	if err != nil {
		log.Fatalf("Failed to load ca file; caFile=%s, err=%v", caFile, err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	tlsConfig.BuildNameToCertificate()

	return func() (net.Conn, error) {
		return tls.Dial("tcp", address, tlsConfig)
	}
}
