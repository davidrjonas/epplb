package main

import (
	"log"
	"net"

	"github.com/davidrjonas/go-epp-proxy/epp"

	pool "gopkg.in/fatih/pool.v2"
)

type ProxyPair struct {
	upstream *pool.PoolConn
	client   *net.TCPConn
}

func (pp *ProxyPair) talk(done func()) {
	defer done()
	defer pp.upstream.Close()

	//pp.upstream.Connect()
	//pp.upstream.Login()

	for {
		frame, err := epp.ReadFrame(pp.client)

		if err != nil {
			log.Printf("Failed to read client; %v", err)
			return
		}

		if frame.IsCommand("login") || frame.IsCommand("logout") {
			if err = epp.WriteFrame(pp.client, frame.MakeSuccessResponse()); err != nil {
				log.Printf("Failed to write client; %v", err)
				return
			}
			continue
		}

		if err = epp.WriteFrame(pp.upstream, frame); err != nil {
			log.Printf("Failed to write upstream; %v", err)

			pp.upstream.MarkUnusable()

			if suberr := epp.WriteFrame(pp.client, frame.MakeErrorResponse(err)); suberr != nil {
				log.Printf("Failed to write client; %v", suberr)
			}

			return
		}

		response, err := epp.ReadFrame(pp.upstream)

		if err != nil {
			log.Printf("Failed to read upstream; %v", err)
			pp.upstream.MarkUnusable()
			response = frame.MakeErrorResponse(err)
		}

		if err := epp.WriteFrame(pp.client, response); err != nil {
			log.Printf("Failed to write client; %v", err)
			return
		}
	}
}
