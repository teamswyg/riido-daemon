package saasplane

import (
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestTaskRequestAppliesSelectedCodexModel(t *testing.T) {
	assignment := assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		ModelID:         "gpt-5.5",
		Prompt:          "do the thing",
		LeaseToken:      "lease-1",
	}
	req := taskRequestFromAssignment(assignment)
	if req.Model != assignment.ModelID {
		t.Fatalf("provider model override = %q, want %q", req.Model, assignment.ModelID)
	}
	if got := req.Metadata[MetadataModelID]; got != assignment.ModelID {
		t.Fatalf("metadata model_id = %q, want %q", got, assignment.ModelID)
	}
}

func TestTaskRequestKeepsSyntheticCodexDefaultAsMetadata(t *testing.T) {
	assignment := assignmentcontract.Assignment{
		RuntimeProvider: "codex",
		ModelID:         "codex-default",
	}
	req := taskRequestFromAssignment(assignment)
	if req.Model != "" {
		t.Fatalf("synthetic default override = %q, want empty", req.Model)
	}
	if got := req.Metadata[MetadataModelID]; got != assignment.ModelID {
		t.Fatalf("metadata model_id = %q, want %q", got, assignment.ModelID)
	}
}

func TestTaskRequestKeepsNonCodexModelOverride(t *testing.T) {
	assignment := assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "claude",
		ModelID:         "claude-3-5-sonnet",
		Prompt:          "do the thing",
		LeaseToken:      "lease-1",
	}
	req := taskRequestFromAssignment(assignment)
	if req.Model != assignment.ModelID {
		t.Fatalf("provider model override = %q, want %q", req.Model, assignment.ModelID)
	}
	if got := req.Metadata[MetadataModelID]; got != assignment.ModelID {
		t.Fatalf("metadata model_id = %q, want %q", got, assignment.ModelID)
	}
}
