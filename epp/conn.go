package epp

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Conn struct {
	io.ReadWriter
}

type Client struct {
	*Conn
	greeting      *Frame
	loginResponse *Frame
}

func NewConn(c io.ReadWriter) *Conn {
	return &Conn{ReadWriter: c}
}

func NewClient(c io.ReadWriter) *Client {
	return &Client{Conn: NewConn(c)}
}

func (c *Conn) ReadFrame() (*Frame, error) {
	header := make([]byte, 4)

	// TODO: handle partial read
	if _, err := c.Read(header); err != nil {
		return nil, err
	}

	size := binary.BigEndian.Uint32(header) - 4

	body := make([]byte, size)

	// TODO: handle partial read
	if _, err := c.Read(body); err != nil {
		return nil, err
	}

	return &Frame{Size: size, Raw: body}, nil
}

func (c *Conn) WriteFrame(frame *Frame) error {
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, frame.Size+4)

	// TODO: handle partial write
	if _, err := c.Write(header); err != nil {
		return err
	}

	if _, err := c.Write(frame.Raw); err != nil {
		return err
	}

	return nil
}

func (c *Client) Connect() (*Frame, error) {
	if c.greeting != nil {
		return c.greeting, nil
	}

	frame, err := c.ReadFrame()

	if err != nil {
		return nil, err
	}

	c.greeting = frame

	return frame, nil
}

func (c *Client) LoginWithFrame(frame *Frame) (*Frame, error) {
	if c.loginResponse != nil {
		return c.loginResponse, nil
	}

	if err := c.WriteFrame(frame); err != nil {
		return nil, err
	}

	response, err := c.ReadFrame()

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

func (c *Client) Login(clID, password, newPassword, clTRID string, svcs, exts []string) (*Frame, error) {
	if c.loginResponse != nil {
		return c.loginResponse, nil
	}

	return c.LoginWithFrame(MakeLoginFrame(clID, password, newPassword, clTRID, svcs, exts))
}

func (c *Client) Logout() error {
	if err := c.WriteFrame(MakeLogoutFrame()); err != nil {
		return err
	}

	return nil
}
