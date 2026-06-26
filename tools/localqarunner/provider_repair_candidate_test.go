package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyProviderEvidencePromotesCandidateAutoRepair(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "provider.json")
	body := `{"status":"partial","providers":[{"id":"openclaw","available":true,"integration_status":"failed","repair":{"class":"local_backend_unavailable","owner":"local_operator","mode":"candidate_auto","summary":"backend unavailable","suggested_command":"brew services start ollama || ollama serve"}}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config{providerEvidence: &path}
	evidence := runEvidence{Status: statusPassed, CoverageStatus: statusPassed}
	if err := applyProviderEvidence("/repo", cfg, &evidence); err != nil {
		t.Fatal(err)
	}
	if len(evidence.OpenRepairs) != 1 {
		t.Fatalf("open repairs=%+v", evidence.OpenRepairs)
	}
	if len(evidence.ClosedLoops) != 1 {
		t.Fatalf("closed loops=%+v", evidence.ClosedLoops)
	}
	assertOpenClawRepairCandidate(t, evidence.ClosedLoops[0])
}

func TestApplyProviderEvidencePromotesClaudeApprovalCandidate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "provider.json")
	body := `{"status":"partial","providers":[{"id":"claude","available":true,"integration_status":"failed","repair":{"class":"provider_tool_approval_missing","owner":"engineer","mode":"candidate_auto","summary":"approval count 0","suggested_command":"verify /tool-approvals is non-empty"}}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config{providerEvidence: &path}
	evidence := runEvidence{Status: statusPassed, CoverageStatus: statusPassed}
	if err := applyProviderEvidence("/repo", cfg, &evidence); err != nil {
		t.Fatal(err)
	}
	if len(evidence.ClosedLoops) != 1 {
		t.Fatalf("closed loops=%+v", evidence.ClosedLoops)
	}
	candidate := evidence.ClosedLoops[0]
	if candidate.ID != "repair-provider.claude.provider_tool_approval_missing" {
		t.Fatalf("candidate id=%q", candidate.ID)
	}
	if candidate.Graph.Observation == "" || candidate.Graph.NextLoop == "" {
		t.Fatalf("candidate graph=%+v", candidate.Graph)
	}
	if len(candidate.RequiredNextArtifacts) == 0 {
		t.Fatalf("candidate artifacts=%+v", candidate.RequiredNextArtifacts)
	}
}

func assertOpenClawRepairCandidate(t *testing.T, candidate runLoopCandidate) {
	t.Helper()
	if candidate.ID != "repair-provider.openclaw.local_backend_unavailable" {
		t.Fatalf("candidate id=%q", candidate.ID)
	}
	if candidate.Graph.Verifier != "local.qa.provider_repair_candidates" {
		t.Fatalf("candidate graph=%+v", candidate.Graph)
	}
	if len(candidate.RequiredNextArtifacts) == 0 {
		t.Fatalf("candidate artifacts=%+v", candidate.RequiredNextArtifacts)
	}
}
