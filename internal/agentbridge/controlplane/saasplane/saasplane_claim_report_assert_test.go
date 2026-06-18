package saasplane

import (
	"strings"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func assertClaimReportRequest(t *testing.T, req *bridge.TaskRequest, assignment assignmentcontract.Assignment) {
	t.Helper()
	if req == nil || req.ID != assignment.ID || string(req.Provider) != assignment.RuntimeProvider {
		t.Fatalf("request = %+v", req)
	}
	if got := req.Metadata[MetadataAssignmentID]; got != assignment.ID {
		t.Fatalf("assignment metadata = %q", got)
	}
	if req.ResumeSessionID != assignment.ResumeSessionID {
		t.Fatalf("resume_session_id = %q", req.ResumeSessionID)
	}
	assertClaimReportWorktree(t, req)
	assertClaimReportPrompt(t, req)
}

func assertClaimReportWorktree(t *testing.T, req *bridge.TaskRequest) {
	t.Helper()
	if req.Worktree == nil ||
		req.Worktree.RepositoryFullName != "teamswyg/riido-daemon" ||
		req.Worktree.BranchName != "RIID-4964-agent-profile-upload" {
		t.Fatalf("worktree = %+v", req.Worktree)
	}
	if got := req.Metadata["workspace_id"]; got != "component-1" {
		t.Fatalf("workspace_id = %q", got)
	}
}

func assertClaimReportPrompt(t *testing.T, req *bridge.TaskRequest) {
	t.Helper()
	if !strings.Contains(req.Prompt, "<riido_log>") ||
		!strings.Contains(req.Prompt, "golang hello world") ||
		!strings.Contains(req.Prompt, "write concise Korean progress updates") {
		t.Fatalf("prompt missing telemetry contract: %q", req.Prompt)
	}
	if got := req.Metadata[agentbridge.MetadataTelemetryContract]; got != agentbridge.TelemetryPlacementPrompt {
		t.Fatalf("telemetry placement = %q", got)
	}
	if got := req.Metadata[agentbridge.MetadataAgentInstruction]; got != agentbridge.TelemetryPlacementPrompt {
		t.Fatalf("instruction placement = %q", got)
	}
}
