package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadJSONRejectsUnknownAndTrailingValues(t *testing.T) {
	dir := t.TempDir()
	valid := filepath.Join(dir, "valid.json")
	if err := os.WriteFile(valid, []byte(`{"schema_version":"v","id":"id"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var detail detailDoc
	if err := readJSON(valid, &detail); err != nil {
		t.Fatal(err)
	}

	unknown := filepath.Join(dir, "unknown.json")
	if err := os.WriteFile(unknown, []byte(`{"schema_version":"v","id":"id","extra":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := readJSON(unknown, &detail); err == nil {
		t.Fatal("expected unknown field to fail")
	}

	trailing := filepath.Join(dir, "trailing.json")
	if err := os.WriteFile(trailing, []byte(`{"schema_version":"v","id":"id"} {}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := readJSON(trailing, &detail); err == nil {
		t.Fatal("expected trailing JSON value to fail")
	}
}

func TestWriteTextCreatesParentDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "doc.md")
	if err := writeText(path, "hello"); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "hello" {
		t.Fatalf("text = %q err=%v", got, err)
	}
}
