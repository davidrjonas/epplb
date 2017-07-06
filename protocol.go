package main

import (
	"errors"

	"github.com/davidrjonas/epplb/epp"
)

type UpstreamError error

type stateFn func() (stateFn, error)

type Protocol struct {
	Upstream   *epp.Client
	Downstream *epp.Conn
}

func (p *Protocol) Talk() (err error) {
	state := p.Connected

	for {
		if state, err = state(); err != nil {
			return err
		}

		if state == nil {
			break
		}
	}

	return nil
}

func (p *Protocol) Connected() (stateFn, error) {
	greeting, err := p.Upstream.Connect()
	if err != nil {
		// TODO: Inform the Downstream? Return a retry?
		return nil, err.(UpstreamError)
	}

	if err := p.Downstream.WriteFrame(greeting); err != nil {
		return nil, err
	}

	return p.Greeted, nil
}

func (p *Protocol) Greeted() (stateFn, error) {
	cmd, err := p.Downstream.ReadFrame()

	if !cmd.IsCommand("login") {
		p.Downstream.WriteFrame(cmd.MakeErrorResponse(errors.New("unauthorized")))
		return p.Greeted, nil
	}

	response, err := p.Upstream.LoginWithFrame(cmd)
	if err != nil {
		// TODO: Inform the Downstream? Return a retry?
		return nil, err.(UpstreamError)
	}

	if err := p.Downstream.WriteFrame(response); err != nil {
		return nil, err
	}

	return p.LoggedIn, nil
}

func (p *Protocol) LoggedIn() (stateFn, error) {
	cmd, err := p.Downstream.ReadFrame()

	if cmd.IsCommand("logout") {
		p.Downstream.WriteFrame(cmd.MakeSuccessResponse())
		return nil, nil
	}

	response, err := p.Upstream.GetResponse(cmd)

	if err != nil {
		// TODO: Inform the Downstream? Return a retry?
		return nil, err.(UpstreamError)

	}

	if err = p.Downstream.WriteFrame(response); err != nil {
		return nil, err
	}

	return p.LoggedIn, nil
}
