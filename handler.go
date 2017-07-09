package main

import (
	"errors"
	"io"
	"log"
	"net"

	pool "gopkg.in/fatih/pool.v2"

	"github.com/davidrjonas/epplb/epp"
)

type ProxyHandler struct {
	pool       pool.Pool
	MaxRetries uint8
}

func (h *ProxyHandler) Handle(c net.Conn) error {
	upstream, err := h.pool.Get()

	if err != nil {
		return err
	}

	p := &Protocol{
		Upstream:   epp.NewClient(upstream),
		Downstream: epp.NewConn(c),
	}

	err = p.Talk()

	if err != nil {
		if _, ok := err.(UpstreamError); ok {
			upstream.(*pool.PoolConn).MarkUnusable()
			log.Printf("upstream error; downstream=%v, upstream=%v, err=%v", c.RemoteAddr(), upstream.RemoteAddr(), err)
			if nErr, ok := err.(RetryableUpstreamError); ok && h.MaxRetries > 0 {
				return h.retryFrame(0, p, nErr.failedFrame)
			}
		} else if err == io.EOF {
			log.Printf("client disconnected; %v", c.RemoteAddr())
			return nil
		}
	}

	return err
}

func (h *ProxyHandler) retryFrame(retryCount uint8, p *Protocol, frame *epp.Frame) error {
	if retryCount >= h.MaxRetries {
		log.Printf("max retries reached; count=%d, downstream=%v, cmd=%v", retryCount, p.Downstream.RemoteAddr(), frame.GetCommand())
		return errors.New("max retries reached")
	}

	log.Printf("retrying failed frame; downstream=%v, cmd=%v", p.Downstream.RemoteAddr(), frame.GetCommand())

	upstream, err := h.pool.Get()

	if err != nil {
		log.Printf("retry failed to get new upstream; downstream=%v, err=%v", p.Downstream.RemoteAddr(), err)
		return err
	}

	p.Upstream = epp.NewClient(upstream)

	err = p.Resume(frame)

	if err != nil {
		if _, ok := err.(UpstreamError); ok {
			upstream.(*pool.PoolConn).MarkUnusable()
			log.Printf("upstream error; downstream=%v, upstream=%v, err=%v", p.Downstream.RemoteAddr(), upstream.RemoteAddr(), err)
			if nErr, ok := err.(RetryableUpstreamError); ok {
				return h.retryFrame(retryCount+1, p, nErr.failedFrame)
			}
		} else if err == io.EOF {
			log.Printf("client disconnected; %v", p.Downstream.RemoteAddr())
			return nil
		}
	}

	return err
}
