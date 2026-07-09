package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildEvidenceValidateAndSourceChecks(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "source.go"), []byte("package x\nmarker"), 0o644); err != nil {
		t.Fatal(err)
	}
	checks := checkSources(dir, []sourceCheck{{Name: "marker", File: "source.go", Contains: "marker"}})
	if len(checks) != 1 || !checks[0].Pass || checks[0].Detail != "" {
		t.Fatalf("checks = %+v", checks)
	}
	failed := checkSources(dir, []sourceCheck{{Name: "missing", File: "source.go", Contains: "nope"}})
	problems := failedChecks("source check failed", failed)
	if len(problems) != 1 || !strings.Contains(problemError(problems).Error(), "source check failed") {
		t.Fatalf("problems = %+v", problems)
	}
	m := manifest{
		ID:           "matrix",
		GeneratedDoc: "root.md",
		DetailDocs: []detailDoc{
			{Path: "a.md"}, {Path: "b.md"}, {Path: "c.md"}, {Path: "d.md"}, {Path: "e.md"},
		},
		ProviderValidation: providerValidation{
			Providers: []providerEvidence{{Provider: "codex", DisplayName: "Codex", OptInIntegration: "manual"}},
		},
	}
	ev := buildEvidence(m, problems, failed)
	if ev.Status != "failed" || ev.ProviderCount != 1 || len(ev.GeneratedDocs) != 6 {
		t.Fatalf("evidence = %+v", ev)
	}
}

func TestValidateManifestReportsRequiredCollections(t *testing.T) {
	problems := validateManifest(manifest{})
	if len(problems) < 3 {
		t.Fatalf("problems = %+v", problems)
	}
}
