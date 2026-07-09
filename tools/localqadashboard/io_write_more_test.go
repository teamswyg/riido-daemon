package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteTextAndJSONCreateNestedFiles(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "a", "dashboard.html")
	if err := writeText(textPath, "<html>ok</html>"); err != nil {
		t.Fatalf("write text: %v", err)
	}
	if data, _ := os.ReadFile(textPath); string(data) != "<html>ok</html>" {
		t.Fatalf("unexpected text output: %q", data)
	}
	jsonPath := filepath.Join(dir, "b", "snapshot.json")
	if err := writeJSON(jsonPath, map[string]any{"ok": true}); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if data, _ := os.ReadFile(jsonPath); len(data) == 0 || data[len(data)-1] != '\n' {
		t.Fatalf("json should be newline-terminated: %q", data)
	}
}
