package main

import (
	"log"
	"net"

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
		frame, err := ReadFrame(pp.client)

		if err != nil {
			log.Printf("Failed to read client; %v", err)
			return
		}

		if frame.IsCommand("login") || frame.IsCommand("logout") {
			if err = WriteFrame(pp.client, frame.MakeSuccessResponse()); err != nil {
				log.Printf("Failed to write client; %v", err)
				return
			}
			continue
		}

		if err = WriteFrame(pp.upstream, frame); err != nil {
			log.Printf("Failed to write upstream; %v", err)

			pp.upstream.MarkUnusable()

			if suberr := WriteFrame(pp.client, frame.MakeErrorResponse(err)); suberr != nil {
				log.Printf("Failed to write client; %v", suberr)
			}

			return
		}

		response, err := ReadFrame(pp.upstream)

		if err != nil {
			log.Printf("Failed to read upstream; %v", err)
			pp.upstream.MarkUnusable()
			response = frame.MakeErrorResponse(err)
		}

		if err := WriteFrame(pp.client, response); err != nil {
			log.Printf("Failed to write client; %v", err)
			return
		}
	}
}
