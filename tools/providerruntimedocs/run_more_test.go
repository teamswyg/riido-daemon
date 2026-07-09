package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesDocsEvidenceAndChecksDrift(t *testing.T) {
	repo := t.TempDir()
	manifestPath := "docs/provider-runtime.riido.json"
	m := testManifest()
	m.Fragments = []string{"claude.riido.json", "codex.riido.json", "cursor.riido.json"}
	m.Pages = nil
	m.SourceChecks = []sourceCheck{{Name: "anchor", File: "source.go", Contains: "runtime anchor"}}
	writeRuntimeManifestFixture(t, repo, manifestPath, m)
	if err := writeText(repoPath(repo, m.Workflow), "workflow"); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
	if err := os.WriteFile(repoPath(repo, "source.go"), []byte("runtime anchor"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	evidenceOut := filepath.Join(repo, "out", "provider-runtime-docs.json")
	opts := options{Repo: repo, Manifest: manifestPath, WriteDoc: true, CheckDoc: true, EvidenceOut: evidenceOut}
	if err := run(opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	body, err := os.ReadFile(evidenceOut)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if !strings.Contains(string(body), `"status": "verified"`) {
		t.Fatalf("evidence body = %s", body)
	}
	if err := os.WriteFile(repoPath(repo, m.GeneratedDoc), []byte("stale"), 0o644); err != nil {
		t.Fatalf("stale doc: %v", err)
	}
	if err := run(options{Repo: repo, Manifest: manifestPath, CheckDoc: true}); err == nil ||
		!strings.Contains(err.Error(), "generated doc drift") {
		t.Fatalf("drift run err = %v", err)
	}
}

func TestRunReportsInvalidManifest(t *testing.T) {
	repo := t.TempDir()
	manifestPath := "docs/provider-runtime.riido.json"
	m := testManifest()
	m.Fragments = []string{"claude.riido.json", "codex.riido.json", "cursor.riido.json"}
	m.Pages = nil
	m.Workflow = "missing.yml"
	m.SourceChecks = []sourceCheck{{Name: "anchor", File: "missing.go", Contains: "runtime"}}
	writeRuntimeManifestFixture(t, repo, manifestPath, m)
	err := run(options{Repo: repo, Manifest: manifestPath})
	if err == nil || !strings.Contains(err.Error(), "missing artifact") ||
		!strings.Contains(err.Error(), "missing.go") {
		t.Fatalf("invalid run err = %v", err)
	}
}

func writeRuntimeManifestFixture(t *testing.T, repo, manifestPath string, m manifest) {
	t.Helper()
	for i, fragment := range m.Fragments {
		p := testPage(strings.TrimSuffix(fragment, ".riido.json"))
		p.GeneratedDoc = "docs/provider-runtime/" + p.ID + ".md"
		if err := writeJSON(fragmentPath(repo, manifestPath, fragment), p); err != nil {
			t.Fatalf("write page %d: %v", i, err)
		}
	}
	if err := writeJSON(repoPath(repo, manifestPath), m); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}
