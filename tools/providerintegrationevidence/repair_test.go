package main

import "testing"

func TestClassifyRepairAuthRequired(t *testing.T) {
	got := classifyRepair("claude", "skipped", "Not logged in", true)
	if got.Class != "provider_auth_required" || got.Owner != "human" {
		t.Fatalf("repair=%+v", got)
	}
}

func TestClassifyRepairOpenClawBackend(t *testing.T) {
	got := classifyRepair("openclaw", "skipped", "openclaw local model backend unavailable", true)
	if got.Class != "local_backend_unavailable" || got.Mode != "candidate_auto" {
		t.Fatalf("repair=%+v", got)
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
