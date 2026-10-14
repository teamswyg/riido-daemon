package saasplane

import (
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestTaskRequestKeepsSyntheticDefaultModelIDAsMetadataOnly(t *testing.T) {
	cases := []struct {
		name     string
		provider string
		modelID  string
	}{
		{name: "codex fallback default", provider: "codex", modelID: "codex-default"},
		{name: "claude fallback default", provider: "claude", modelID: "claude-default"},
		{name: "openclaw fallback default", provider: "openclaw", modelID: "openclaw-default"},
		{name: "cursor auto default", provider: "cursor", modelID: "cursor-auto"},
		{name: "unknown fallback default", provider: "other", modelID: "runtime-default"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assignment := assignmentcontract.Assignment{
				ID:              "asn-1",
				TaskID:          "task-a",
				ComponentID:     "component-1",
				AgentID:         "jykim1",
				RuntimeProvider: tc.provider,
				ModelID:         tc.modelID,
				Prompt:          "do the thing",
				LeaseToken:      "lease-1",
			}
			req := taskRequestFromAssignment(assignment)
			if req.Model != "" {
				t.Fatalf("provider model override = %q, want empty for synthetic default", req.Model)
			}
			if got := req.Metadata[MetadataModelID]; got != tc.modelID {
				t.Fatalf("metadata model_id = %q, want %q", got, tc.modelID)
			}
		})
	}
}
