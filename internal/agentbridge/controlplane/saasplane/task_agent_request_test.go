package saasplane

import (
	"strings"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestTaskAgentRequestsKeepExecutionIdentityAndUserInstruction(t *testing.T) {
	first := assignmentcontract.Assignment{ID: "assignment-one", TaskID: "task", AgentID: "agent", WorkspaceID: "workspace", RequestID: "request-one", RuntimeProvider: "codex", Prompt: "Task title\n\n[user request] fix the bug", AgentInstruction: "Follow repository rules"}
	second := first
	second.ID = "assignment-two"
	second.RequestID = "request-two"
	a, b := taskRequestFromAssignment(first), taskRequestFromAssignment(second)
	if a.ID == b.ID || a.Metadata["run_id"] == b.Metadata["run_id"] {
		t.Fatal("different requests collapsed")
	}
	if a.Metadata["workspace_id"] != "workspace" || a.Metadata["request_id"] != "request-one" {
		t.Fatalf("missing request metadata: %#v", a.Metadata)
	}
	if !strings.Contains(a.Prompt, "fix the bug") || strings.Contains(a.SystemPrompt, "fix the bug") {
		t.Fatal("user request crossed the system prompt boundary")
	}
	if retry := taskRequestFromAssignment(first); retry.ID != a.ID {
		t.Fatal("retry changed execution identity")
	}
	legacy := first
	legacy.WorkspaceID = ""
	legacy.RequestID = ""
	legacy.ComponentID = "component"
	if taskRequestFromAssignment(legacy).Metadata["workspace_id"] != "component" {
		t.Fatal("legacy fallback changed")
	}
}

func TestReferenceFoldersReachBothProvidersWithoutChangingCwd(t *testing.T) {
	for _, provider := range []string{"codex", "claude"} {
		assignment := assignmentcontract.Assignment{ID: "assignment", RuntimeProvider: provider, Prompt: "Read relevant material.", AgentInstruction: "Keep my rules", ReferenceFolders: &assignmentcontract.ReferenceFolders{DeviceID: "device", RuntimeID: "runtime", Paths: []string{"/자료 with spaces"}}}
		req := taskRequestFromAssignment(assignment)
		if req.Cwd != "" || req.ReferenceFolders.Paths[0] != "/자료 with spaces" {
			t.Fatalf("%s: bad reference propagation", provider)
		}
		if !strings.Contains(req.Prompt+req.SystemPrompt, "Keep my rules") {
			t.Fatalf("%s: dropped user rules", provider)
		}
		assignment.ReferenceFolders.Paths[0] = "/changed"
		if req.ReferenceFolders.Paths[0] != "/자료 with spaces" {
			t.Fatal("assignment aliased")
		}
	}
}
