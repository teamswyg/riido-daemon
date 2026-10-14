package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRunVerifiesRequiredEvidence(t *testing.T) {
	root := t.TempDir()
	manifestPath := writeFixtureManifest(t, root)
	evidenceDir := filepath.Join(root, "out")
	mustMkdir(t, evidenceDir)
	mustWrite(t, filepath.Join(evidenceDir, "loop.json"), `{"status":"verified","problem_count":0}`)
	docPath := filepath.Join(root, "self.md")
	m, err := loadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeDoc(docPath, m); err != nil {
		t.Fatal(err)
	}
	reportPath := filepath.Join(root, "report.json")
	err = run(options{
		Manifest:    manifestPath,
		EvidenceDir: evidenceDir,
		CheckDoc:    true,
		EvidenceOut: reportPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	report := readReport(t, reportPath)
	if report.ClosedVerified != 2 {
		t.Fatalf("closed loops = %d, want 2", report.ClosedVerified)
	}
}

func TestRunFailsMissingEvidence(t *testing.T) {
	root := t.TempDir()
	manifestPath := writeFixtureManifest(t, root)
	m, err := loadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeDoc(filepath.Join(root, "self.md"), m); err != nil {
		t.Fatal(err)
	}
	err = run(options{
		Manifest:    manifestPath,
		EvidenceDir: filepath.Join(root, "out"),
		CheckDoc:    true,
		EvidenceOut: filepath.Join(root, "report.json"),
	})
	if err == nil {
		t.Fatal("expected missing evidence failure")
	}
	report := readReport(t, filepath.Join(root, "report.json"))
	joined := strings.Join(report.Problems, "\n")
	if !strings.Contains(joined, "go run ./tools/loopevidence") {
		t.Fatalf("missing producer command in %q", joined)
	}
	if !strings.Contains(joined, "loop.json") {
		t.Fatalf("missing expected evidence file in %q", joined)
	}
}
