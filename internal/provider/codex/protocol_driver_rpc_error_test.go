package codex

import (
	"context"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexProtocolDriverRPCErrorFailsPendingRequest(t *testing.T) {
	d, io := startedProtocolDriver(t)

	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      int64(1),
		"error":   map[string]any{"message": "missing auth"},
	}
	events, _, err := d.OnRaw(context.Background(), agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    "error",
		Payload: payload,
	}, io)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("expected failed result for rpc error, got %+v", events)
	}
	if !strings.Contains(last.Result.Error, "initialize") || !strings.Contains(last.Result.Error, "missing auth") {
		t.Fatalf("missing rpc error context: %+v", last.Result)
	}
}
