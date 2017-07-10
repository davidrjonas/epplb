package epp

import (
	"encoding/binary"
	"net"
)

type Conn struct {
	net.Conn
}

func NewConn(c net.Conn) *Conn {
	return &Conn{Conn: c}
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
