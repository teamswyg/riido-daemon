package claude

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildProviderInputApprovalResponse(t *testing.T) {
	body, err := BuildProviderInput(agentbridge.Command{
		Kind:              agentbridge.CommandApproveTool,
		ToolID:            "tu_1",
		ProviderRequestID: "r_1",
	})
	if err != nil {
		t.Fatalf("BuildProviderInput approve: %v", err)
	}
	assertProviderInputContains(t, string(body), []string{
		`"type":"control_response"`,
		`"request_id":"r_1"`,
		`"behavior":"allow"`,
		`"updatedInput":{}`,
	})
}

func TestBuildProviderInputDenyResponse(t *testing.T) {
	body, err := BuildProviderInput(agentbridge.Command{
		Kind:              agentbridge.CommandRejectTool,
		ProviderRequestID: "r_2",
		Reason:            "No shell access",
	})
	if err != nil {
		t.Fatalf("BuildProviderInput deny: %v", err)
	}
	assertProviderInputContains(t, string(body), []string{
		`"request_id":"r_2"`,
		`"behavior":"deny"`,
		`"message":"No shell access"`,
	})
}

func TestBuildProviderInputRequiresProviderRequestID(t *testing.T) {
	if _, err := BuildProviderInput(agentbridge.Command{Kind: agentbridge.CommandApproveTool, ToolID: "tu_1"}); err == nil {
		t.Fatal("expected missing provider request id to fail")
	}
}

func assertProviderInputContains(t *testing.T, raw string, wants []string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(raw, want) {
			t.Fatalf("provider input missing %s: %s", want, raw)
		}
	}
}
