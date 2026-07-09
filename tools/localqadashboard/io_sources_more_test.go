package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProviderAndCoverageEvidenceAttachSources(t *testing.T) {
	dir := t.TempDir()
	providerPath := filepath.Join(dir, "provider.json")
	manifestPath := filepath.Join(dir, "manifest.json")
	writeTestJSON(t, providerPath, providerEvidenceFile{
		ExpiresAt: "2999-01-01T00:00:00Z",
		Providers: []providerEvidence{{ID: "codex", IntegrationStatus: "passed"}},
	})
	writeTestJSON(t, manifestPath, coverageManifest{
		ID: "coverage", Scenarios: []coverageScenario{{ID: "provider.codex"}},
	})
	provider, err := loadProviderEvidence(providerPath)
	if err != nil {
		t.Fatalf("load provider: %v", err)
	}
	if provider.EvidenceArtifact != providerPath {
		t.Fatalf("provider artifact missing: %#v", provider)
	}
	if provider.Providers[0].EvidenceArtifact != providerPath ||
		provider.Providers[0].ExpiresAt == "" {
		t.Fatalf("provider row provenance missing: %#v", provider.Providers[0])
	}
	manifest, err := loadCoverageManifest(manifestPath)
	if err != nil || manifest.ID != "coverage" || len(manifest.Scenarios) != 1 {
		t.Fatalf("load manifest: manifest=%#v err=%v", manifest, err)
	}
}

func TestLoadExternalEvidenceHandlesMissingAndProvenance(t *testing.T) {
	if got, err := loadExternalEvidence(""); err != nil || got.ID != "" {
		t.Fatalf("empty external path should be optional: %#v %v", got, err)
	}
	if got, err := loadExternalEvidence(filepath.Join(t.TempDir(), "missing.json")); err != nil || got.ID != "" {
		t.Fatalf("missing external path should be optional: %#v %v", got, err)
	}
	path := filepath.Join(t.TempDir(), "external.json")
	writeTestJSON(t, path, externalEvidenceFile{
		ID: "external", ExpiresAt: "2999-01-01T00:00:00Z",
		Scenarios: []externalScenario{{ID: "contract.task.thread_message", Status: "passed"}},
	})
	got, err := loadExternalEvidence(path)
	if err != nil {
		t.Fatalf("load external: %v", err)
	}
	if got.Evidence != path || got.Scenarios[0].Evidence != path ||
		got.Scenarios[0].ExpiresAt == "" {
		t.Fatalf("external provenance missing: %#v", got)
	}
}

func writeTestJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
