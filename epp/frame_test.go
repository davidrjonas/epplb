package epp

import (
	"errors"
	"strings"
	"testing"
)

var xml_command_info = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>` +
	`<epp xmlns="urn:ietf:params:xml:ns:epp-1.0"><command><info><obj:info xmlns:obj="urn:ietf:params:xml:ns:obj"><obj:name>example</obj:name></obj:info></info>` +
	`<clTRID>ABC-12345</clTRID></command></epp>`
var xml_command_info_len = len(xml_command_info)

var xml_response_success = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>` +
	`<epp xmlns="urn:ietf:params:xml:ns:epp-1.0"><response><result code="1000"><msg lang="en">Command completed successfully</msg></result>` +
	`<trID><clTRID>ABC-12345</clTRID><svTRID>54321-XYZ</svTRID></trID></response></epp>`

func TestFrameFromString(t *testing.T) {
	f := FrameFromString(xml_command_info)
	if string(f.Raw) != xml_command_info {
		t.Error("Frame.Raw does not match provided xml")
	}

	if int(f.Size) != xml_command_info_len {
		t.Errorf("Expected %v, got %v", xml_command_info_len, f.Size)
	}
}

func TestIsCommandHandlesCommand(t *testing.T) {
	f := FrameFromString(xml_command_info)

	if !f.IsCommand("info") {
		t.Error("Expected true")
	}

	if f.IsCommand("notinfo") {
		t.Error("Expected false")
	}
}

func TestIsCommandHandlesNonCommand(t *testing.T) {
	f := FrameFromString(xml_response_success)

	if f.IsCommand("response") {
		t.Error("Expected false")
	}

	if f.IsCommand("result") {
		t.Error("Expected false")
	}
}

func TestGetCommand(t *testing.T) {
	f := FrameFromString(xml_command_info)
	cmd := f.GetCommand()

	if cmd != "info" {
		t.Error("Expected 'info', got", cmd)
	}

}

func TestGetClTRIDForCommand(t *testing.T) {
	f := FrameFromString(xml_command_info)
	clTRID := f.GetClTRID()

	if clTRID != "ABC-12345" {
		t.Error("Expected 'ABC-12345', got", clTRID)
	}
}

func TestGetClTRIDForResponse(t *testing.T) {
	f := FrameFromString(xml_response_success)
	clTRID := f.GetClTRID()

	if clTRID != "ABC-12345" {
		t.Error("Expected 'ABC-12345', got", clTRID)
	}
}

func TestMakeSuccessResponseIncludesClTRID(t *testing.T) {
	f := FrameFromString(xml_command_info)
	clTRID := f.GetClTRID()

	res := f.MakeSuccessResponse()
	if !strings.Contains(string(res.Raw), clTRID) {
		t.Error("Response should contain clTRID ", clTRID)
	}
}

func TestMakeErrorResponseIncludesErrorMsgAndClTRID(t *testing.T) {
	f := FrameFromString(xml_command_info)
	clTRID := f.GetClTRID()

	res := f.MakeErrorResponse(errors.New("foo"))

	if !strings.Contains(string(res.Raw), clTRID) {
		t.Error("Response should contain clTRID ", clTRID)
	}
	if !strings.Contains(string(res.Raw), "<msg>foo</msg>") {
		t.Error("Response should contain msg foo")
	}
}

func TestGetResultProvidesCodeAndMsg(t *testing.T) {
	f := FrameFromString(xml_response_success)
	res, err := f.GetResult()
	if err != nil {
		t.Errorf("GetResult failed with %v", err)
	}

	if res.Code != 1000 {
		t.Errorf("Expected 1000, got %v", res.Code)
	}

	if res.Msg != "Command completed successfully" {
		t.Errorf("Expected 'Command completed successfully', got %v", res.Msg)
	}
}

func TestGetResultErrorOnNonResponse(t *testing.T) {
	f := FrameFromString(xml_command_info)
	_, err := f.GetResult()
	if err == nil {
		t.Error("Expected error was nil")
	}
	if err.Error() != "frame is not a response" {
		t.Error("Expected error 'frame is not a response', got ", err.Error())
	}
}

func TestGetResultErrorOnMissingResult(t *testing.T) {
	var xml_invalid_response = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>` +
		`<epp xmlns="urn:ietf:params:xml:ns:epp-1.0"><response><trID><clTRID>ABC-12345</clTRID><svTRID>54321-XYZ</svTRID></trID></response></epp>`
	f := FrameFromString(xml_invalid_response)
	_, err := f.GetResult()
	if err == nil {
		t.Error("Expected error was nil")
	}
	if err.Error() != "frame is missing result" {
		t.Error("Expected error 'frame is missing result', got ", err.Error())
	}
}

func TestIsSuccess(t *testing.T) {
	// Success codes
	for _, code := range []string{"1000", "1001", "1500", "1999"} {
		f := FrameFromString(strings.Replace(xml_response_success, "1000", code, 1))
		if !f.IsSuccess() {
			t.Error("Expected success for code", code)
		}
	}
	// Failure codes
	for _, code := range []string{"2000", "2001", "2500", "2999", "3000"} {
		f := FrameFromString(strings.Replace(xml_response_success, "1000", code, 1))
		if f.IsSuccess() {
			t.Error("Expected failure for code", code)
		}
	}
}

func TestIsFailure(t *testing.T) {
	// Failure codes
	for _, code := range []string{"2000", "2001", "2500", "2999", "3000"} {
		f := FrameFromString(strings.Replace(xml_response_success, "1000", code, 1))
		if !f.IsFailure() {
			t.Error("Expected failure for code", code)
		}
	}
	// Success codes
	for _, code := range []string{"1000", "1001", "1500", "1999"} {
		f := FrameFromString(strings.Replace(xml_response_success, "1000", code, 1))
		if f.IsFailure() {
			t.Error("Expected success for code", code)
		}
	}
}
