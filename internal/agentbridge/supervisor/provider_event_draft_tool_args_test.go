package supervisor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestProviderEventDraftIncludesToolArgs(t *testing.T) {
	_, payload, ok := providerEventDraft(agentbridge.Event{
		Kind: agentbridge.EventToolCallStarted,
		Tool: agentbridge.ToolRef{
			ID:   "tool-1",
			Name: "Bash",
			Kind: "shell",
			Args: map[string]string{"command": "go test ./..."},
		},
	})
	if !ok {
		t.Fatal("expected tool event mapping")
	}

	args, ok := payload["args"].(map[string]string)
	if !ok {
		t.Fatalf("args payload type = %T", payload["args"])
	}
	if args["command"] != "go test ./..." {
		t.Fatalf("args payload = %+v", args)
	}
}
