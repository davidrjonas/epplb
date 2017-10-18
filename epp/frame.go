package epp

import (
	"fmt"
	"log"
	"strings"

	"github.com/jteeuwen/go-pkg-xmlx"
)

const nsEpp10 = "urn:ietf:params:xml:ns:epp-1.0"

type Frame struct {
	Size uint32
	Raw  []byte
	doc  *xmlx.Document
}

type Result struct {
	Code uint16
	Msg  string
}

func FrameFromString(xml string) *Frame {
	b := []byte(xml)
	return &Frame{Raw: b, Size: uint32(len(b))}
}

func (f *Frame) IsCommand(cmd string) bool {
	return f.GetCommand() == cmd
}

func (f *Frame) GetCommand() string {
	doc := f.getDoc()
	node := doc.SelectNode(nsEpp10, "command")

	if node == nil {
		return ""
	}

	for _, child := range node.Children {
		if child.Type == xmlx.NT_ELEMENT {
			return child.Name.Local
		}
	}

	return ""
}

func (f *Frame) GetClTRID() string {
	doc := f.getDoc()
	node := doc.SelectNode(nsEpp10, "command")
	if node != nil {
		return node.S(nsEpp10, "clTRID")
	}

	node = doc.SelectNode(nsEpp10, "trID")

	if node != nil {
		return node.S(nsEpp10, "clTRID")
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

func (f *Frame) GetResult() (*Result, error) {
	doc := f.getDoc()
	node := doc.SelectNode(nsEpp10, "response")

	if node == nil || len(node.Children) == 0 {
		return nil, fmt.Errorf("frame is not a response")
	}

	result := node.SelectNode(nsEpp10, "result")

	if result == nil {
		return nil, fmt.Errorf("frame is missing result")
	}

	code := result.Au16("", "code")
	msg := result.S(nsEpp10, "msg")

	return &Result{Code: code, Msg: msg}, nil
}

func (f *Frame) IsSuccess() bool {
	result, err := f.GetResult()
	if err != nil {
		return false
	}

	return result.Code >= 1000 && result.Code < 2000
}

func (f *Frame) IsFailure() bool {
	result, err := f.GetResult()
	if err != nil {
		return false
	}

	return result.Code >= 2000
}

// MakeLoginFrame is not implemented
//func MakeLoginFrame(clID, password, newPassword, clTRID string, svcs, exts []string) *Frame {
//	panic(errors.New("not implemented"))
//	return &Frame{}
//}

func MakeHelloFrame() *Frame {
	xml := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>\n` +
		`<epp xmlns="urn:ietf:params:xml:ns:epp-1.0"><hello/></epp>`)
	return &Frame{Raw: xml, Size: uint32(len(xml))}
}

func MakeLogoutFrame() *Frame {
	xml := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>\n` +
		`<epp xmlns="urn:ietf:params:xml:ns:epp-1.0"><command><logout/><clTRID>00000-AAA</clTRID></command></epp>`)
	return &Frame{Raw: xml, Size: uint32(len(xml))}
}
