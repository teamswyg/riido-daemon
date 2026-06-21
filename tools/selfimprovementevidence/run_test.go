package main

import (
	"path/filepath"
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
	err = run(options{
		Manifest:    manifestPath,
		EvidenceDir: evidenceDir,
		CheckDoc:    true,
		EvidenceOut: filepath.Join(root, "report.json"),
	})
	if err != nil {
		t.Fatal(err)
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
}

func writeFixtureManifest(t *testing.T, root string) string {
	t.Helper()
	path := filepath.Join(root, "manifest.json")
	mustWrite(t, path, `{
  "schema_version":"riido-daemon-self-improvement-evidence.v1",
  "id":"fixture",
  "title":"Fixture",
  "generated_doc":"`+filepath.Join(root, "self.md")+`",
  "workflow":".github/workflows/self-improvement-evidence.yml",
  "evidence_artifact":"self-improvement-evidence",
  "loop_source":"docs/30-architecture/loop-engineering/self-improvement-evidence.riido.json",
  "required_evidence":[{"id":"loop","file":"loop.json","description":"loop","assertions":[{"field":"status","equals":"verified"},{"field":"problem_count","equals":0}]}]
}`)
	return path
}
