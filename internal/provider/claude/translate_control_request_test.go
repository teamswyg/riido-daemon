package claude

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateControlRequestProducesApprovalNeeded(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"control_request","request_id":"r_1","request":{"subtype":"permission_request","tool_use_id":"tu_1","tool_name":"Bash","tool_input":{"command":"terraform destroy"}}}`)
	events := translate(t, raw)

	if len(events) != 1 {
		t.Fatalf("want 1 event, got %d: %+v", len(events), events)
	}
	if events[0].Kind != agentbridge.EventToolApprovalNeeded {
		t.Fatalf("control_request must produce approval event, got %s", events[0].Kind)
	}
	if events[0].Tool.ID != "tu_1" || events[0].Tool.Name != "Bash" {
		t.Fatalf("tool ref: %+v", events[0].Tool)
	}
	if events[0].Tool.ProviderRequestID != "r_1" {
		t.Fatalf("provider request id: %+v", events[0].Tool)
	}
	if events[0].Tool.Args["command"] != "terraform destroy" {
		t.Fatalf("tool args: %+v", events[0].Tool.Args)
	}
}
