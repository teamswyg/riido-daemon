package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeJSONRejectsUnknownAndTrailingValues(t *testing.T) {
	var m manifest
	if err := decodeJSON([]byte(`{"schema_version":"v","id":"id"}`), &m); err != nil {
		t.Fatal(err)
	}
	if err := decodeJSON([]byte(`{"schema_version":"v","id":"id","extra":true}`), &m); err == nil {
		t.Fatal("expected unknown field to fail")
	}
	if err := decodeJSON([]byte(`{"schema_version":"v","id":"id"} {}`), &m); err == nil {
		t.Fatal("expected trailing JSON value to fail")
	}
}

func TestLoadEntryFilesAndWriteEvidence(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")
	entryPath := filepath.Join(dir, "entries.json")
	body := `[{"node_id":"n1","name":"Screen","upstream_owner":["figma"],` +
		`"daemon_scope":"runtime","daemon_consumed_facts":[],"client_owned_facts":["copy"]}]`
	if err := os.WriteFile(entryPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := loadEntryFiles(manifestPath, []string{"entries.json"})
	if err != nil || len(entries) != 1 || entries[0].NodeID != "n1" {
		t.Fatalf("entries=%+v err=%v", entries, err)
	}
	out := filepath.Join(dir, "nested", "evidence.json")
	if err := writeJSON(out, map[string]string{"status": "verified"}); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(out); err != nil || string(got) != "{\n  \"status\": \"verified\"\n}\n" {
		t.Fatalf("json = %q err=%v", got, err)
	}
}
