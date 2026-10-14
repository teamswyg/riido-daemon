package claude

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type controlIO struct {
	closed bool
	writes [][]byte
}

func (io *controlIO) WriteStdin(_ context.Context, b []byte) error {
	if io.closed {
		return errors.New("stdin closed")
	}
	io.writes = append(io.writes, append([]byte(nil), b...))
	return nil
}
func (io *controlIO) CloseStdin(context.Context) error { io.closed = true; return nil }

func TestClaudeApprovalRetainsInputAndOpenControlChannel(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{Prompt: "perform task"})
	io := &controlIO{}
	if err := d.OnStart(context.Background(), io); err != nil {
		t.Fatal(err)
	}
	if io.closed {
		t.Fatal("control stream closed before approval")
	}
	raw := mustParseRaw(t, `{"type":"control_request","request_id":"req-1","request":{"subtype":"can_use_tool","tool_use_id":"write-1","tool_name":"Write","input":{"file_path":"/tmp/result.txt","content":"exact content","secret":"retained only for provider transport"}}}`)
	events, _, err := d.OnRaw(context.Background(), raw, io)
	if err != nil || len(events) != 1 || events[0].Kind != agentbridge.EventToolApprovalNeeded {
		t.Fatalf("events=%+v err=%v", events, err)
	}
	if events[0].Tool.Args["secret"] != "[redacted]" {
		t.Fatal("raw secret entered diagnostic tool args")
	}
	builder := d.(agentbridge.ProviderInputBuilder)
	frame, err := builder.BuildProviderInput(agentbridge.Command{Kind: agentbridge.CommandApproveTool, ProviderRequestID: "req-1"})
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(frame, &decoded); err != nil {
		t.Fatal(err)
	}
	response := mapField(mapField(decoded, "response"), "response")
	input := mapField(response, "updatedInput")
	if input["file_path"] != "/tmp/result.txt" || input["content"] != "exact content" || input["secret"] != "retained only for provider transport" {
		t.Fatalf("native input changed: %+v", input)
	}
	if err := io.WriteStdin(context.Background(), frame); err != nil {
		t.Fatal(err)
	}
	if _, err := builder.BuildProviderInput(agentbridge.Command{Kind: agentbridge.CommandApproveTool, ProviderRequestID: "req-1"}); err == nil {
		t.Fatal("completed control request retained")
	}
	_ = d.OnClose(context.Background(), io)
}
