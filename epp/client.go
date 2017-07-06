package epp

import (
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	conn              *Conn
	busy              sync.Mutex
	lastOp            int64
	keepaliveInterval time.Duration
	keepaliveTicker   *time.Ticker
	greeting          *Frame
	loginResponse     *Frame
}

type ClientOption func(*Client)

func KeepaliveInterval(d time.Duration) ClientOption {
	return func(c *Client) {
		c.keepaliveInterval = d
	}
}

func NewClient(c io.ReadWriter, options ...ClientOption) *Client {
	client := Client{
		conn:              NewConn(c),
		keepaliveInterval: 5 * time.Minute,
	}

	for _, opt := range options {
		opt(&client)
	}

	client.keepaliveStart()

	return &client
}

func (c *Client) readFrame() (*Frame, error) {
	return c.conn.ReadFrame()
}

func (c *Client) writeFrame(f *Frame) error {
	atomic.StoreInt64(&c.lastOp, time.Now().UnixNano())
	return c.conn.WriteFrame(f)
}

func (c *Client) keepaliveStart() {
	if c.keepaliveInterval <= 0 {
		return
	}

	ticker := time.NewTicker(c.keepaliveInterval)
	go func() {
		for t := range ticker.C {
			lastOp := time.Unix(0, atomic.LoadInt64(&c.lastOp))
			if t.After(lastOp.Add(c.keepaliveInterval)) {
				log.Println("sending keepalive; lastOp=" + lastOp.Format(time.RFC3339))
				c.Hello()
			}
		}
	}()

	c.keepaliveTicker = ticker
}

func (c *Client) keepaliveStop() {
	if c.keepaliveTicker == nil {
		return
	}

	c.keepaliveTicker.Stop()
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
	c.keepaliveStop()

	if _, err := c.GetResponse(MakeLogoutFrame()); err != nil {
		return err
	}

	return nil
}
