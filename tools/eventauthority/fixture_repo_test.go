package main

import (
	"os"
	"path/filepath"
	"testing"
)

func testRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	dir := filepath.Join(repo, "internal", "ir", "ingest")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(dir, "draft.go"), draftSource())
	mustWrite(t, filepath.Join(dir, "event_from_draft.go"), builderSource())
	return repo
}

func draftSource() string {
	return `package ingest
type Draft struct {
	OccurredAt string
	Scope string
	Type string
	Payload map[string]any
	Unknown map[string]any
	TaskID string
}`
}

func builderSource() string {
	return `package ingest
func build(draft Draft) ir.CanonicalEvent {
	return ir.CanonicalEvent{EventID: "id", EventSchemaVersion: "v1", ActorKind: "daemon", ActorID: "actor", OccurredAt: draft.OccurredAt, Scope: draft.Scope, Type: draft.Type, Payload: draft.Payload, Unknown: draft.Unknown, TaskID: draft.TaskID}
}`
}
