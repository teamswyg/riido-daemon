package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReportsEvidenceWriteFailure(t *testing.T) {
	dir, manifestPath := fixture(t)
	err := run(dir, manifestPath, "missing/evidence.json", true, false)
	if err == nil || !strings.Contains(err.Error(), "write evidence") {
		t.Fatalf("expected evidence write error, got %v", err)
	}
}

func TestLoadManifestReportsReadAndParseFailures(t *testing.T) {
	_, err := loadManifest(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil || !strings.Contains(err.Error(), "read manifest") {
		t.Fatalf("expected read manifest error, got %v", err)
	}
	path := filepath.Join(t.TempDir(), "bad.json")
	mustWrite(t, path, "{")
	_, err = loadManifest(path)
	if err == nil || !strings.Contains(err.Error(), "parse manifest") {
		t.Fatalf("expected parse manifest error, got %v", err)
	}
}

func TestMaybeWriteOrCheckReportsMissingGeneratedDoc(t *testing.T) {
	dir, manifestPath := fixture(t)
	loaded, err := loadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "docs", "README.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "docs", "readme", "document-map.md")); err != nil {
		t.Fatal(err)
	}
	err = maybeWriteOrCheck(dir, loaded, false, true)
	if err == nil || !strings.Contains(err.Error(), "read generated doc") {
		t.Fatalf("expected missing generated doc error, got %v", err)
	}
}

func TestValidateManifestReportsDuplicateTopicAndMissingFields(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "docs", "README.md"), "")
	mustWrite(t, filepath.Join(dir, "docs", "readme", "document-map.md"), "")
	mustWrite(t, filepath.Join(dir, "docs", "a.md"), "")
	m := manifest{
		SchemaVersion: "bad",
		GeneratedDocs: generatedDocs{
			Readme:      "docs/README.md",
			DocumentMap: "docs/readme/document-map.md",
		},
		ReadOrder: []readEntry{{Doc: "docs/a.md"}},
		Decisions: []decision{{Topic: "A", Docs: []string{"docs/a.md"}}, {Topic: "A"}},
		Repos:     []repo{{Repo: "r", Responsibility: "x"}},
		Rules:     []string{"rule"},
	}
	problems := strings.Join(validateManifest(dir, m), "\n")
	for _, want := range []string{"schema_version", "required", "duplicate decision topic", "require doc"} {
		if !strings.Contains(problems, want) {
			t.Fatalf("problems missing %q: %s", want, problems)
		}
	}
}
