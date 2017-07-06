package epp

import (
	"fmt"
	"io"
	"sync"
)

type Client struct {
	conn          *Conn
	busy          sync.Mutex
	greeting      *Frame
	loginResponse *Frame
}

func NewClient(c io.ReadWriter) *Client {
	return &Client{conn: NewConn(c)}
}

func (c *Client) readFrame() (*Frame, error) {
	return c.conn.ReadFrame()
}

func (c *Client) writeFrame(f *Frame) error {
	return c.conn.WriteFrame(f)
}

func (c *Client) Connect() (*Frame, error) {
	if c.greeting != nil {
		return c.greeting, nil
	}

	frame, err := c.readFrame()

	if err != nil {
		return nil, err
	}

	c.greeting = frame

	return frame, nil
}

func (c *Client) GetResponse(f *Frame) (*Frame, error) {
	c.busy.Lock()
	defer c.busy.Unlock()

	err := c.writeFrame(f)

	if err != nil {
		return nil, err
	}

	response, err := c.readFrame()

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) LoginWithFrame(frame *Frame) (*Frame, error) {
	if c.loginResponse != nil {
		return c.loginResponse, nil
	}

	response, err := c.GetResponse(frame)

	if err != nil {
		return nil, err
	}

	if response.IsFailure() {
		result, err := response.GetResult()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("login failed; %v", result.Msg)
	}

	c.loginResponse = response

	return response, nil
}

func (c *Client) Hello() (*Frame, error) {
	return c.GetResponse(MakeHelloFrame())
}

//func (c *Client) Login(clID, password, newPassword, clTRID string, svcs, exts []string) (*Frame, error) {
//	if c.loginResponse != nil {
//		return c.loginResponse, nil
//	}
//
//	return c.LoginWithFrame(MakeLoginFrame(clID, password, newPassword, clTRID, svcs, exts))
//}

func (c *Client) Logout() error {
	if err := c.writeFrame(MakeLogoutFrame()); err != nil {
		return err
	}

	// TODO: do we need to read a response?

	return nil
}
