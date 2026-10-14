package saasplane

import (
	"strings"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTaskRequestPlacesTelemetryForSystemPromptProviders(t *testing.T) {
	assignment := assignmentcontract.Assignment{
		ID:                       "asn-1",
		TaskID:                   "task-a",
		ComponentID:              "component-1",
		AgentID:                  "jykim1",
		RuntimeProvider:          "claude",
		ModelID:                  "claude-opus-4-7",
		Prompt:                   "golang hello world quickly",
		AgentInstruction:         "act as a backend reviewer",
		AllowExperimentalRuntime: true,
		ResumeSessionID:          "sess-prev",
	}
	req := taskRequestFromAssignment(assignment)
	if req.Prompt != assignment.Prompt {
		t.Fatalf("claude prompt should remain user task only: %q", req.Prompt)
	}
	if !strings.Contains(req.SystemPrompt, "<riido_log>") ||
		!strings.Contains(req.SystemPrompt, "act as a backend reviewer") ||
		!strings.Contains(req.SystemPrompt, "어떤 작업부터 진행할까요?") {
		t.Fatalf("claude system prompt missing runtime instructions: %q", req.SystemPrompt)
	}
	if got := req.Metadata[agentbridge.MetadataTelemetryContract]; got != agentbridge.TelemetryPlacementSystemPrompt {
		t.Fatalf("telemetry placement = %q", got)
	}
	if got := req.Metadata[agentbridge.MetadataAgentInstruction]; got != agentbridge.TelemetryPlacementSystemPrompt {
		t.Fatalf("instruction placement = %q", got)
	}
	if !req.AllowExperimentalRuntime || req.Model != assignment.ModelID || req.ResumeSessionID != assignment.ResumeSessionID {
		t.Fatalf("task request lost assignment fields: %+v", req)
	}
	if got := req.Metadata[MetadataModelID]; got != assignment.ModelID {
		t.Fatalf("metadata model_id = %q, want %q", got, assignment.ModelID)
	}
}
