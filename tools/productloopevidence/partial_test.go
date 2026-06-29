package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBuildPartialReductionComputesCandidateAge(t *testing.T) {
	now := time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)
	m := manifest{Thresholds: thresholds{StalePartialAfterDays: 7}}
	reg := registrySource{Loops: []registryLoop{
		{ID: "fresh", Kind: "systematization-audit", CandidateCreatedAt: "2026-06-24", PromotionTarget: "verifier"},
		{ID: "stale", Kind: "systematization-audit", CandidateCreatedAt: "2026-06-01", PromotionTarget: "gate"},
		{ID: "unknown", Kind: "systematization-audit"},
		{ID: "closed", Kind: "closed-loop"},
	}}
	got := buildPartialReductionAt(t.TempDir(), m, reg, qaSystemSource{}, now)
	if got.CandidateAgeUnknownCount != 1 || got.StaleCandidateCount != 1 {
		t.Fatalf("partial reduction = %+v", got)
	}
	if len(got.CandidateAges) != 2 || got.CandidateAges[0].AgeDays != 2 || !got.CandidateAges[1].Stale {
		t.Fatalf("candidate ages = %+v", got.CandidateAges)
	}
}

func TestBuildPartialReductionRequiresFreshLocalQARunEvidence(t *testing.T) {
	root := t.TempDir()
	rel := ".riido-local/evidence/local-qa-run.json"
	now := time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)
	m := manifest{
		LocalQARunEvidence: rel,
		Thresholds:         thresholds{StalePartialAfterDays: 7},
	}
	writeRunEvidence(t, root, rel, "2000-01-01T00:00:00Z")
	got := buildPartialReductionAt(root, m, registrySource{}, qaSystemSource{}, now)
	if got.LocalQARunEvidenceState != localQARunExpired || got.LocalQARunEvidenceFresh {
		t.Fatalf("expired evidence treated as fresh: %+v", got)
	}
	writeRunEvidence(t, root, rel, "2999-01-01T00:00:00Z")
	got = buildPartialReductionAt(root, m, registrySource{}, qaSystemSource{}, now)
	if got.LocalQARunEvidenceState != localQARunFresh || !got.LocalQARunEvidenceFresh {
		t.Fatalf("fresh evidence not accepted: %+v", got)
	}
}

func writeRunEvidence(t *testing.T, root, rel, expires string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"expires_at":"` + expires + `"}`)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
}
