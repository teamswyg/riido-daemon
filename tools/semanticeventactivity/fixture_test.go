package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func validManifest(doc string) Manifest {
	return Manifest{
		SchemaVersion:       schemaVersion,
		ID:                  "test",
		Title:               "Test",
		GeneratedDoc:        doc,
		Workflow:            "workflow.yml",
		EvidenceArtifact:    "artifact",
		SemanticActivity:    []string{"lifecycle", "text_delta", "thinking_delta", "tool_call_started", "tool_call_delta", "tool_call_completed", "tool_call_failed", "tool_approval_needed", "usage_delta", "progress"},
		NonSemanticActivity: []string{"session_identified", "log", "warning", "error", "result", "cancellation_requested", "timeout", "process_exit"},
		Assertions:          []string{"classification matches runtime"},
	}
}

func mustWriteManifest(t *testing.T, repo, path string, manifest Manifest) {
	t.Helper()
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(repo, path), string(data))
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func testRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	dir := filepath.Join(repo, "internal", "agentbridge")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(dir, "event_kind.go"), eventKindSource())
	return repo
}
