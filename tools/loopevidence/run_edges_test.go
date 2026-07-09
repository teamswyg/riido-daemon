package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesDocAndEvidenceForLoopFile(t *testing.T) {
	dir, manifestPath := writeManifestWithLoopFile(t)
	evidencePath := filepath.Join(dir, "out", "loop-evidence.json")

	if err := run(options{Repo: dir, Manifest: manifestPath, Write: true, EvidenceOut: evidencePath}); err != nil {
		t.Fatalf("run write: %v", err)
	}
	assertFileContains(t, filepath.Join(dir, "loop.md"), "OK")
	assertFileContains(t, evidencePath, `"status": "verified"`)
}

func TestRunCheckReportsGeneratedDocDriftWithEvidence(t *testing.T) {
	dir, manifestPath := writeManifestWithLoopFile(t)
	docPath := filepath.Join(dir, "loop.md")
	if err := os.WriteFile(docPath, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	evidencePath := filepath.Join(dir, "out", "loop-evidence.json")

	err := run(options{Repo: dir, Manifest: manifestPath, Check: true, EvidenceOut: evidencePath})
	if err == nil || !strings.Contains(err.Error(), "generated doc drift") {
		t.Fatalf("expected drift error, got %v", err)
	}
	assertFileContains(t, evidencePath, `"status": "failed"`)
}

func TestLoadManifestReportsReadAndParseErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := loadManifest(filepath.Join(dir, "missing.json")); err == nil ||
		!strings.Contains(err.Error(), "read manifest") {
		t.Fatalf("expected read manifest error, got %v", err)
	}
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadManifest(bad); err == nil || !strings.Contains(err.Error(), "parse manifest") {
		t.Fatalf("expected parse manifest error, got %v", err)
	}
}

func writeManifestWithLoopFile(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	loopPath := filepath.Join(dir, "loops", "x.riido.json")
	writeLoopFixture(t, loopPath)
	manifestPath := filepath.Join(dir, "loop.riido.json")
	data := `{"schema_version":"riido-loop-evidence.v1","id":"ok","title":"OK","generated_doc":"loop.md","required_phases":["observe","hypothesis","execute","evaluate","retrospective"],"loop_files":["loops/x.riido.json"],"open_gaps":[]}`
	if err := os.WriteFile(manifestPath, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, manifestPath
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("%s does not contain %q: %s", path, want, string(data))
	}
}
