package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCoverageEvidenceReportsReadAndParseErrors(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.json")
	cfg := config{coverageEvidence: &missing}
	if err := applyCoverageEvidence(".", cfg, &runEvidence{}); err == nil ||
		!strings.Contains(err.Error(), "read coverage evidence") {
		t.Fatalf("expected coverage read error, got %v", err)
	}
	bad := filepath.Join(dir, "coverage.json")
	mustRunnerWrite(t, bad, "{")
	cfg.coverageEvidence = &bad
	if err := applyCoverageEvidence(".", cfg, &runEvidence{}); err == nil ||
		!strings.Contains(err.Error(), "parse coverage evidence") {
		t.Fatalf("expected coverage parse error, got %v", err)
	}
}

func TestProviderEvidenceReportsReadAndParseErrors(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.json")
	cfg := config{providerEvidence: &missing}
	if err := applyProviderEvidence(".", cfg, &runEvidence{}); err == nil ||
		!strings.Contains(err.Error(), "read provider evidence") {
		t.Fatalf("expected provider read error, got %v", err)
	}
	bad := filepath.Join(dir, "provider.json")
	mustRunnerWrite(t, bad, "{")
	cfg.providerEvidence = &bad
	if err := applyProviderEvidence(".", cfg, &runEvidence{}); err == nil ||
		!strings.Contains(err.Error(), "parse provider evidence") {
		t.Fatalf("expected provider parse error, got %v", err)
	}
}

func TestProductEvidenceIsOptionalButRejectsMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.json")
	cfg := config{productEvidence: &missing}
	evidence := runEvidence{CoverageStatus: statusPassed}
	if err := applyProductEvidence(".", cfg, &evidence); err != nil {
		t.Fatalf("missing product evidence should be optional: %v", err)
	}
	if evidence.CoverageStatus != statusPassed {
		t.Fatalf("coverage changed for missing product evidence: %+v", evidence)
	}
	bad := filepath.Join(dir, "product.json")
	mustRunnerWrite(t, bad, "{")
	cfg.productEvidence = &bad
	if err := applyProductEvidence(".", cfg, &runEvidence{}); err == nil ||
		!strings.Contains(err.Error(), "parse product evidence") {
		t.Fatalf("expected product parse error, got %v", err)
	}
}

func mustRunnerWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
