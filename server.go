package main

import (
	"log"
	"net"
	"sync"
	"time"
)

type Handler func(net.Conn) error

type Server struct {
	listener      *net.TCPListener
	acceptTimeout time.Duration
	stop          chan bool
	done          chan bool
}

func NewServer(listener net.Listener) *Server {
	return &Server{
		listener:      listener.(*net.TCPListener),
		acceptTimeout: 10 * time.Millisecond,
		stop:          make(chan bool),
		done:          make(chan bool, 1),
	}
}

func (s *Server) Stop() {
	close(s.stop)
	<-s.done
}

func (s *Server) Serve(handle Handler) {
	var wg sync.WaitGroup

OUTER:
	for {
		s.listener.SetDeadline(time.Now().Add(s.acceptTimeout))

		conn, err := s.listener.Accept()

		if err != nil {
			if opErr, ok := err.(*net.OpError); !ok || !opErr.Timeout() {
				log.Printf("error accepting connection: %v", err)
			}

			select {
			case <-s.stop:
				break OUTER
			default:
				continue OUTER
			}
		}

		wg.Add(1)

		go func(handle Handler, conn net.Conn, done func()) {
			if err := handle(conn); err != nil {
				log.Printf("connection error: %v", err)
			}
			conn.Close()
			done()
		}(handle, conn, wg.Done)
	}

	log.Println("waiting for clients to finish")

	wg.Wait()
	s.done <- true
}
