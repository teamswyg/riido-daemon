package main

import (
	"strings"
	"testing"
)

func TestClassifyRepairAuthRequired(t *testing.T) {
	got := classifyRepair("claude", "skipped", "Not logged in", true)
	if got.Class != "provider_auth_required" || got.Owner != "human" {
		t.Fatalf("repair=%+v", got)
	}
}

func TestClassifyRepairCursorAuthIncludesAction(t *testing.T) {
	got := classifyRepair("cursor", "skipped", "account missing", true)
	if !strings.Contains(got.SuggestedCommand, "cursor-agent login") {
		t.Fatalf("repair=%+v", got)
	}
	if !strings.Contains(got.SuggestedCommand, "CURSOR_API_KEY") {
		t.Fatalf("repair=%+v", got)
	}
}

func TestClassifyRepairOpenClawBackend(t *testing.T) {
	got := classifyRepair("openclaw", "skipped", "openclaw local model backend unavailable", true)
	if got.Class != "local_backend_unavailable" || got.Mode != "candidate_auto" {
		t.Fatalf("repair=%+v", got)
	}
}

func TestClassifyRepairClaudeToolApprovalMissing(t *testing.T) {
	got := classifyRepair("claude", "failed", "go run command execution permission limited; approval count 0", true)
	if got.Class != "provider_tool_approval_missing" || got.Mode != "candidate_auto" || got.Owner != "engineer" {
		t.Fatalf("repair=%+v", got)
	}
	if !strings.Contains(got.SuggestedCommand, "/tool-approvals") {
		t.Fatalf("repair must point to approval evidence: %+v", got)
	}
}

func TestClassifyRepairClaudeConversationalApprovalGap(t *testing.T) {
	got := classifyRepair("claude", "failed", "go run 명령어 실행 권한이 제한되어 있습니다. 대화창에서 승인해 주세요.", true)
	if got.Class != "provider_tool_approval_missing" || got.Mode != "candidate_auto" {
		t.Fatalf("repair=%+v", got)
	}
	if !strings.Contains(got.SuggestedCommand, "chat approval") {
		t.Fatalf("repair must require chat approval binding evidence: %+v", got)
	}
}

func TestClassifyRepairOpenClawCooldown(t *testing.T) {
	got := classifyRepair("openclaw", "failed", "Provider ollama is in cooldown; All models failed", true)
	if got.Class != "local_backend_unavailable" || got.Owner != "local_operator" {
		t.Fatalf("repair=%+v", got)
	}
}

func TestClassifyRepairMissingExecutable(t *testing.T) {
	got := classifyRepair("cursor", "skipped", "executable not found", false)
	if got.Class != "provider_executable_missing" {
		t.Fatalf("repair=%+v", got)
	}
}

func TestClassifyRepairProviderTimeout(t *testing.T) {
	got := classifyRepair("codex", "failed", "Error:hard timeout", true)
	if got.Class != "provider_timeout" || got.Owner != "engineer" {
		t.Fatalf("repair=%+v", got)
	}
}

func TestClassifyRepairSideEffectMissing(t *testing.T) {
	got := classifyRepair("codex", "failed", "completed without writing expected artifact", true)
	if got.Class != "provider_side_effect_missing" || got.Owner != "engineer" {
		t.Fatalf("repair=%+v", got)
	}
}

func TestClassifyRepairOpenClawSideEffectUsesModelConfig(t *testing.T) {
	got := classifyRepair("openclaw", "failed", "completed without writing expected artifact", true)
	if got.Class != "openclaw_cwd_side_effect_unverified" || got.Owner != "local_operator" {
		t.Fatalf("repair=%+v", got)
	}
}
