package epp

import (
	"encoding/binary"
	"io"
	"net"
)

type Conn struct {
	io.ReadWriter
}

func NewConn(c io.ReadWriter) *Conn {
	return &Conn{ReadWriter: c}
}

func (c *Conn) RemoteAddr() net.Addr {
	panic("not implemented")
	return nil
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
