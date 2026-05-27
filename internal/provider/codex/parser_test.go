package codex

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func feedAll(t *testing.T, p agentbridge.Parser, chunks ...string) []agentbridge.RawEvent {
	t.Helper()
	var raws []agentbridge.RawEvent
	for _, c := range chunks {
		r, err := p.FeedStdout([]byte(c))
		if err != nil {
			t.Fatalf("FeedStdout %q: %v", c, err)
		}
		raws = append(raws, r...)
	}
	closed, err := p.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
	raws = append(raws, closed...)
	return raws
}

// JSON-RPC notification: no id, has method + params.
func TestParserNotification(t *testing.T) {
	raws := feedAll(t, NewParser(), `{"jsonrpc":"2.0","method":"agent_message","params":{"text":"hi"}}`+"\n")
	if len(raws) != 1 {
		t.Fatalf("want 1 raw, got %d", len(raws))
	}
	if raws[0].Type != "notification:agent_message" {
		t.Fatalf("type: %q", raws[0].Type)
	}
}

// JSON-RPC response: id present, result present, no method.
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

// JSON-RPC error response.
func TestParserErrorResponse(t *testing.T) {
	raws := feedAll(t, NewParser(), `{"jsonrpc":"2.0","id":2,"error":{"code":-32000,"message":"boom"}}`+"\n")
	if len(raws) != 1 || raws[0].Type != "error" {
		t.Fatalf("err response: %+v", raws)
	}
}

// JSON-RPC server-initiated request: id + method.
func TestParserServerRequest(t *testing.T) {
	raws := feedAll(t, NewParser(), `{"jsonrpc":"2.0","id":3,"method":"approve_command","params":{"command":"rm -rf /"}}`+"\n")
	if len(raws) != 1 {
		t.Fatalf("want 1 raw, got %d", len(raws))
	}
	if raws[0].Type != "server_request:approve_command" {
		t.Fatalf("type: %q", raws[0].Type)
	}
}

func TestParserPartialLineReassembly(t *testing.T) {
	p := NewParser()
	raws := feedAll(t, p, `{"jsonrpc":"2.0","method":"x`, `"}`+"\n")
	if len(raws) != 1 {
		t.Fatalf("reassembly: %+v", raws)
	}
}

func TestParserMalformedNonFatal(t *testing.T) {
	chunk := "garbage\n" + `{"jsonrpc":"2.0","method":"ok"}` + "\n"
	raws := feedAll(t, NewParser(), chunk)
	if len(raws) != 2 {
		t.Fatalf("want 2, got %d", len(raws))
	}
	if raws[0].Type != "malformed" {
		t.Fatalf("first must be malformed: %+v", raws[0])
	}
	if raws[1].Type != "notification:ok" {
		t.Fatalf("second must be ok: %+v", raws[1])
	}
}

func TestParserStderrTagged(t *testing.T) {
	p := NewParser()
	r, _ := p.FeedStderr([]byte("warn line\n"))
	if len(r) != 1 {
		t.Fatalf("want 1, got %d", len(r))
	}
	if r[0].Source != agentbridge.RawSourceStderr || r[0].Type != "stderr" {
		t.Fatalf("source/type: %+v", r[0])
	}
	if !strings.Contains(string(r[0].Bytes), "warn line") {
		t.Fatalf("bytes: %q", r[0].Bytes)
	}
}
