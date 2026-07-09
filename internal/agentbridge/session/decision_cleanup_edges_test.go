package session

import (
	"os"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestToolDecisionNilGatesAllowByDefault(t *testing.T) {
	tool := agentbridge.ToolRef{ID: "tool-1", Name: "write"}
	if got := decideStartedTool(nil, tool); got != (agentbridge.ToolStartDecision{}) {
		t.Fatalf("started nil gate decision = %+v", got)
	}
	if got := decideApprovalTool(nil, tool); got != (agentbridge.ToolStartDecision{}) {
		t.Fatalf("approval nil gate decision = %+v", got)
	}
}

func TestToolDecisionDelegatesGates(t *testing.T) {
	tool := agentbridge.ToolRef{ID: "tool-1", Name: "write"}
	want := agentbridge.ToolStartDecision{Block: true, Code: "blocked"}
	gate := func(got agentbridge.ToolRef) agentbridge.ToolStartDecision {
		if got.ID != tool.ID || got.Name != tool.Name {
			t.Fatalf("tool = %+v", got)
		}
		return want
	}
	if got := decideStartedTool(gate, tool); got != want {
		t.Fatalf("started decision = %+v", got)
	}
	if got := decideApprovalTool(agentbridge.ToolApprovalGate(gate), tool); got != want {
		t.Fatalf("approval decision = %+v", got)
	}
}

func TestToolDecisionReasonOnly(t *testing.T) {
	got := toolBlockReason(agentbridge.ToolStartDecision{Reason: "unsafe path"})
	if got != "unsafe path" {
		t.Fatalf("reason-only block = %q", got)
	}
}

func TestCleanupTempFilesSkipsEmptyMissingAndDuplicates(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/scratch"
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if got := cleanupTempFiles([]string{"", path, path, dir + "/missing"}); len(got) != 0 {
		t.Fatalf("cleanup events = %+v", got)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("temp file still exists, err=%v", err)
	}
}

func TestCleanupTempFilesReportsRemoveFailureOnce(t *testing.T) {
	dir := t.TempDir()
	child := dir + "/child"
	if err := os.WriteFile(child, []byte("x"), 0o600); err != nil {
		t.Fatalf("write child: %v", err)
	}
	got := cleanupTempFiles([]string{dir, dir})
	if len(got) != 1 {
		t.Fatalf("events len=%d, events=%+v", len(got), got)
	}
	if got[0].Kind != agentbridge.EventWarning {
		t.Fatalf("event kind=%s", got[0].Kind)
	}
	if !strings.Contains(got[0].Err, dir) {
		t.Fatalf("event err=%q, want path", got[0].Err)
	}
}
