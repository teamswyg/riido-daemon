package codex

import (
	"encoding/json"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildProviderInputApprovalResponse(t *testing.T) {
	body, err := BuildProviderInput(agentbridge.Command{
		Kind:              agentbridge.CommandApproveTool,
		ProviderRequestID: "77",
	})
	if err != nil {
		t.Fatalf("BuildProviderInput approve: %v", err)
	}
	frame := decodeProviderInput(t, body)
	if frame["jsonrpc"] != "2.0" || frame["id"] != float64(77) {
		t.Fatalf("approval frame: %+v", frame)
	}
	result := frame["result"].(map[string]any)
	if result["decision"] != "accept" {
		t.Fatalf("approval decision: %+v", result)
	}
}

func TestBuildProviderInputRejectionCancelsTurn(t *testing.T) {
	body, err := BuildProviderInput(agentbridge.Command{
		Kind:              agentbridge.CommandRejectTool,
		ProviderRequestID: "req-1",
	})
	if err != nil {
		t.Fatalf("BuildProviderInput reject: %v", err)
	}
	frame := decodeProviderInput(t, body)
	if frame["id"] != "req-1" {
		t.Fatalf("rejection frame: %+v", frame)
	}
	result := frame["result"].(map[string]any)
	if result["decision"] != "cancel" {
		t.Fatalf("rejection decision: %+v", result)
	}
}

func TestBuildProviderInputRequiresProviderRequestID(t *testing.T) {
	if _, err := BuildProviderInput(agentbridge.Command{Kind: agentbridge.CommandApproveTool}); err == nil {
		t.Fatal("expected missing provider request id to fail")
	}
}

func decodeProviderInput(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var frame map[string]any
	if err := json.Unmarshal(body, &frame); err != nil {
		t.Fatalf("decode provider input: %v; body=%s", err, string(body))
	}
	return frame
}
