package main

import (
	"errors"

	"github.com/davidrjonas/epplb/epp"
)

type UpstreamError error

type RetryableUpstreamError struct {
	UpstreamError
	failedFrame *epp.Frame
}

type stateFn func() (stateFn, error)

type Protocol struct {
	Upstream   *epp.Client
	Downstream *epp.Conn
}

func (p *Protocol) Talk() (err error) {
	return p.run(p.connected)
}

func (p *Protocol) Resume(f *epp.Frame) error {
	if f == nil {
		return p.run(p.connected)
	}

	switch f.GetCommand() {
	case "login":
		stateFn, err := p.greetedThenFrame(f)
		if err != nil {
			return err
		}
		return p.run(stateFn)
	case "logout":
		p.Downstream.WriteFrame(f.MakeSuccessResponse())
		return nil
	default:
		stateFn, err := p.loggedInThenFrame(f)
		if err != nil {
			return err
		}
		return p.run(stateFn)
	}
}

func (p *Protocol) run(state stateFn) (err error) {
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

func (p *Protocol) connected() (stateFn, error) {
	greeting, err := p.Upstream.Connect()
	if err != nil {
		return nil, RetryableUpstreamError{UpstreamError: err}
	}

	if err := p.Downstream.WriteFrame(greeting); err != nil {
		return nil, err
	}

	return p.greeted, nil
}

func (p *Protocol) greeted() (stateFn, error) {
	cmd, err := p.Downstream.ReadFrame()

	if err != nil {
		return nil, err
	}

	if !cmd.IsCommand("login") {
		p.Downstream.WriteFrame(cmd.MakeErrorResponse(errors.New("unauthorized")))
		return p.greeted, nil
	}

	return p.greetedThenFrame(cmd)
}

func (p *Protocol) greetedThenFrame(cmd *epp.Frame) (stateFn, error) {

	response, err := p.Upstream.LoginWithFrame(cmd)
	if err != nil {
		return nil, RetryableUpstreamError{
			UpstreamError: err,
			failedFrame:   cmd,
		}
	}

	if err := p.Downstream.WriteFrame(response); err != nil {
		return nil, err
	}

	return p.loggedIn, nil
}

func (p *Protocol) loggedIn() (stateFn, error) {
	cmd, err := p.Downstream.ReadFrame()

	if err != nil {
		return nil, err
	}

	if cmd.IsCommand("logout") {
		p.Downstream.WriteFrame(cmd.MakeSuccessResponse())
		return nil, nil
	}

	return p.loggedInThenFrame(cmd)
}

func (p *Protocol) loggedInThenFrame(cmd *epp.Frame) (stateFn, error) {

	response, err := p.Upstream.GetResponse(cmd)

	if err != nil {
		return nil, RetryableUpstreamError{
			UpstreamError: err,
			failedFrame:   cmd,
		}
	}

	if err = p.Downstream.WriteFrame(response); err != nil {
		return nil, err
	}

	return p.loggedIn, nil
}
