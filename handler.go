package main

import (
	"net"

	pool "gopkg.in/fatih/pool.v2"

	"github.com/davidrjonas/go-epp-proxy/epp"
)

type ProxyHandler struct {
	pool pool.Pool
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
		}
	}

	return err
}
