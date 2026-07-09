package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestDocAndEvidenceEdges(t *testing.T) {
	dir, manifestPath := fixture(t)
	if _, err := loadManifest(filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("expected missing manifest error")
	}
	mustWrite(t, filepath.Join(dir, "bad.json"), "{")
	if _, err := loadManifest(filepath.Join(dir, "bad.json")); err == nil {
		t.Fatal("expected malformed manifest error")
	}
	loaded, err := loadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	loaded.SchemaVersion = "bad"
	loaded.Workflow = "missing.yml"
	loaded.Commands = nil
	loaded.Assertions = nil
	problems := validateManifest(dir, loaded)
	for _, want := range []string{"schema_version", "missing workflow", "commands must not be empty", "assertions"} {
		assertRepoProblem(t, problems, want)
	}
	if err := maybeWriteOrCheckDoc(dir, loaded, false, false); err != nil {
		t.Fatal(err)
	}
	loaded.GeneratedDoc = "missing-doc.md"
	assertRepoError(t, maybeWriteOrCheckDoc(dir, loaded, false, true), "read generated doc")
	loaded.GeneratedDoc = "doc.md"
	assertRepoError(t, maybeWriteOrCheckDoc(dir, loaded, false, true), "generated doc drift")
	if err := maybeWriteOrCheckDoc(dir, loaded, true, true); err != nil {
		t.Fatal(err)
	}
	if err := maybeWriteOrCheckDoc(dir, loaded, false, true); err != nil {
		t.Fatal(err)
	}
}

func TestEvidenceAndRunFailureEdges(t *testing.T) {
	dir, manifestPath := fixture(t)
	assertRepoError(t, writeJSON(filepath.Join(dir, "bad.json"), map[string]any{"bad": make(chan int)}), "encode evidence")
	assertRepoError(t, writeText(filepath.Join(dir, "missing", "doc.md"), "x"), "write generated doc")
	bad := filepath.Join(dir, "bad-evidence")
	if err := os.Mkdir(bad, 0o755); err != nil {
		t.Fatal(err)
	}
	assertRepoError(t, run(dir, manifestPath, bad, false, false, false), "write evidence")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	mutated := strings.Replace(string(data), `["echo","a"]`, `["sh","-c","exit 2"]`, 1)
	mustWrite(t, manifestPath, mutated)
	assertRepoError(t, run(dir, manifestPath, "", false, false, true), "one or more verification commands failed")
	ev := buildEvidence(manifest{ID: "m", Assertions: []string{"a"}}, []commandEvidence{{Status: "failed"}})
	if ev.Status != "failed" || ev.ID != "m" || len(ev.Assertions) != 1 {
		t.Fatalf("evidence=%#v", ev)
	}
}
