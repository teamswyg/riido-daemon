package codex

import "testing"

func TestParserNotification(t *testing.T) {
	raws := feedAll(t, NewParser(), `{"jsonrpc":"2.0","method":"agent_message","params":{"text":"hi"}}`+"\n")
	if len(raws) != 1 {
		t.Fatalf("want 1 raw, got %d", len(raws))
	}
	if raws[0].Type != "notification:agent_message" {
		t.Fatalf("type: %q", raws[0].Type)
	}
}

func TestParserResponse(t *testing.T) {
	raws := feedAll(t, NewParser(), `{"jsonrpc":"2.0","id":1,"result":{"thread_id":"t1"}}`+"\n")
	if len(raws) != 1 {
		t.Fatalf("want 1 raw, got %d", len(raws))
	}
	if raws[0].Type != "response" {
		t.Fatalf("type: %q", raws[0].Type)
	}
	if raws[0].Payload["id"] != float64(1) {
		t.Fatalf("id lost: %+v", raws[0].Payload)
	}
}

func TestParserErrorResponse(t *testing.T) {
	raws := feedAll(t, NewParser(), `{"jsonrpc":"2.0","id":2,"error":{"code":-32000,"message":"boom"}}`+"\n")
	if len(raws) != 1 || raws[0].Type != "error" {
		t.Fatalf("err response: %+v", raws)
	}
}

func TestParserServerRequest(t *testing.T) {
	raws := feedAll(t, NewParser(), `{"jsonrpc":"2.0","id":3,"method":"approve_command","params":{"command":"rm -rf /"}}`+"\n")
	if len(raws) != 1 {
		t.Fatalf("want 1 raw, got %d", len(raws))
	}
	if raws[0].Type != "server_request:approve_command" {
		t.Fatalf("type: %q", raws[0].Type)
	}
}
