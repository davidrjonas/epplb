package epp

import "testing"

var xml_command_info = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>` +
	`<epp xmlns="urn:ietf:params:xml:ns:epp-1.0"><command><info><obj:info xmlns:obj="urn:ietf:params:xml:ns:obj"><obj:name>example</obj:name></obj:info></info>` +
	`<clTRID>ABC-12345</clTRID></command></epp>`

var xml_response_success = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>` +
	`<epp xmlns="urn:ietf:params:xml:ns:epp-1.0"><response><result code="1000"><msg lang="en">Command completed successfully</msg></result>` +
	`<trID><clTRID>ABC-12345</clTRID><svTRID>54321-XYZ</svTRID></trID></response></epp>`

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
