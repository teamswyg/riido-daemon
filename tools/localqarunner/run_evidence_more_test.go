package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewEvidenceCapturesArtifactsAndStrictCoverage(t *testing.T) {
	validFor := time.Hour
	strict := true
	cfg := uploadTestConfig("product.json", "release.json", "coverage.json", "lab.html", "schedule.json", "infra.json")
	cfg.validFor = &validFor
	cfg.strictCoverage = &strict
	cfg.promotionManifest = strPtr("promotions.json")
	cfg.manualEvidence = strPtr("manual.json")
	cfg.domainCache = strPtr("domain.json")
	observed := time.Date(2026, 7, 9, 1, 2, 3, 0, time.UTC)
	ev := newEvidence(cfg, observed)
	if ev.SchemaVersion != "riido-local-qa-run-result.v1" || ev.Status != statusPassed {
		t.Fatalf("unexpected evidence: %#v", ev)
	}
	if !ev.StrictCoverage || ev.ExpiresAt != observed.Add(time.Hour).Format(time.RFC3339) {
		t.Fatalf("strict/expiry missing: %#v", ev)
	}
	if ev.Artifacts.ProductEvidence != "product.json" ||
		ev.Artifacts.PromotionRegistry != "promotions.json" {
		t.Fatalf("artifacts missing: %#v", ev.Artifacts)
	}
}

func TestFinishRunWritesEvidenceAndDeploymentGate(t *testing.T) {
	dir := t.TempDir()
	runPath := filepath.Join("out", "run.json")
	promotionPath := filepath.Join("out", "missing-promotions.json")
	cfg := config{runEvidence: &runPath, promotionManifest: &promotionPath}
	ev := runEvidence{
		SchemaVersion:  "riido-local-qa-run-result.v1",
		ID:             "local-qa-run",
		Status:         statusPassed,
		CoverageStatus: statusPassed,
	}
	status, err := finishRun(dir, cfg, ev)
	if err != nil || status != statusPassed {
		t.Fatalf("finish run: status=%s err=%v", status, err)
	}
	data, err := os.ReadFile(filepath.Join(dir, runPath))
	if err != nil {
		t.Fatalf("read run evidence: %v", err)
	}
	var decoded runEvidence
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode run evidence: %v", err)
	}
	if decoded.DeploymentGate.Status != deploymentGateReady {
		t.Fatalf("deployment gate missing: %#v", decoded.DeploymentGate)
	}
}

func TestAppendStepMarksRunFailed(t *testing.T) {
	ev := runEvidence{Status: statusPassed}
	appendStep(&ev, stepEvidence{ID: "ok", Status: statusPassed})
	appendStep(&ev, stepEvidence{ID: "bad", Status: statusFailed})
	if ev.Status != statusFailed || len(ev.Steps) != 2 {
		t.Fatalf("unexpected step aggregation: %#v", ev)
	}
}

func strPtr(value string) *string {
	return &value
}
