package main

import (
	"encoding/binary"
	"log"
	"net"
	"strings"

	"github.com/jteeuwen/go-pkg-xmlx"
)

const NS_EPP10 = "urn:ietf:params:xml:ns:epp-1.0"

type Frame struct {
	Size uint32
	Raw  []byte
	doc  *xmlx.Document
}

func FromString(xml string) *Frame {
	b := []byte(xml)
	return &Frame{Raw: b, Size: uint32(len(b))}
}

func (f *Frame) IsCommand(cmd string) bool {
	return f.GetCommand() == cmd
}

func (f *Frame) GetCommand() string {
	doc := f.getDoc()
	node := doc.SelectNode(NS_EPP10, "command")

	if node == nil || len(node.Children) == 0 {
		return ""
	}

	return node.Children[0].Name.Local
}

func (f *Frame) GetClTRID() string {
	doc := f.getDoc()
	node := doc.SelectNode(NS_EPP10, "command")
	if node != nil {
		return node.S(NS_EPP10, "clTRID")
	}

	node = doc.SelectNode(NS_EPP10, "trID")

	if node != nil {
		return node.S(NS_EPP10, "clTRID")
	}

	return ""
}

func (f *Frame) getDoc() *xmlx.Document {
	if f.doc != nil {
		return f.doc
	}

	f.doc = xmlx.New()

	if err := f.doc.LoadBytes(f.Raw, nil); err != nil {
		// TODO: handle failure
		log.Printf("Failed to parse xml; %v", err)
		return nil
	}

	return f.doc
}

func (f *Frame) MakeSuccessResponse() *Frame {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="no"?>\n` +
		`<epp xmlns="urn:ietf:params:xml:ns:epp-1.0"><response>` +
		`<result code="1000"><msg>Command completed successfully</msg></result>` +
		`<trID><clTRID>{{clTRID}}</clTRID><svTRID>00000-ZZZ</svTRID></trID>` +
		`</response></epp>`

	b := []byte(strings.Replace(xml, "{{clTRID}}", f.GetClTRID(), 1))
	return &Frame{Raw: b, Size: uint32(len(b))}
}

func (f *Frame) MakeErrorResponse(err error) *Frame {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="no"?>\n` +
		`<epp xmlns="urn:ietf:params:xml:ns:epp-1.0"><response>` +
		`<result code="2400"><msg>{{msg}}</msg></result>` +
		`<trID><clTRID>{{clTRID}}</clTRID><svTRID>00000-ZZZ</svTRID></trID>` +
		`</response></epp>`

	s := strings.Replace(xml, "{{msg}}", err.Error(), 1)
	b := []byte(strings.Replace(s, "{{clTRID}}", f.GetClTRID(), 1))

	return &Frame{Raw: b, Size: uint32(len(b))}
}

func ReadFrame(c net.Conn) (*Frame, error) {
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

	return &Frame{Size: size - 4, Raw: body}, nil
}

func WriteFrame(c net.Conn, frame *Frame) error {
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
