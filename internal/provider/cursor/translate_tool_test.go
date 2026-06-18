package cursor

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateToolUse(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"go test ./...","password":"raw","note":"sk-ant-`+strings.Repeat("a", 24)+`"}}`))
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventToolCallStarted {
		t.Fatalf("tool_use: %+v", evs)
	}
	if evs[0].Tool.Args["command"] != "go test ./..." {
		t.Fatalf("tool args: %+v", evs[0].Tool.Args)
	}
	if evs[0].Tool.Args["password"] != "[redacted]" {
		t.Fatalf("sensitive args must be redacted: %+v", evs[0].Tool.Args)
	}
	if evs[0].Tool.Args["note"] != "[redacted]" {
		t.Fatalf("secret-looking value must be redacted: %+v", evs[0].Tool.Args)
	}
}

func TestTranslateToolResult(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"tool_result","tool_use_id":"t1"}`))
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventToolCallCompleted {
		t.Fatalf("tool_result: %+v", evs)
	}
}
