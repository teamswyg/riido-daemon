package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestRejectsUnknownAndTrailingJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	valid := `{"schema_version":"v","id":"x","title":"X","generated_doc":"doc.md","workflow":"wf.yml","evidence_artifact":"e","purpose":"p","facts":[{"name":"f","summary":"s","source_checks":["src"]}],"source_checks":[{"name":"src","file":"src.go","contains":"needle"}],"assertions":[]}`
	cases := []struct {
		name string
		body string
		want string
	}{
		{name: "unknown", body: strings.Replace(valid, `"id":"x"`, `"id":"x","extra":true`, 1), want: "unknown field"},
		{name: "trailing", body: valid + `{"id":"extra"}`, want: "trailing JSON value"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustNativeWrite(t, path, tc.body)
			_, err := loadManifest(dir, "manifest.json")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}

func TestWritersCreateParentDirectories(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "nested", "evidence.json")
	if err := writeJSON(jsonPath, Evidence{ID: "native"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"id": "native"`) {
		t.Fatalf("json output missing id: %s", data)
	}
	textPath := filepath.Join(dir, "nested", "doc.md")
	if err := writeText(textPath, "doc"); err != nil {
		t.Fatal(err)
	}
	if data, err = os.ReadFile(textPath); err != nil || string(data) != "doc" {
		t.Fatalf("text output=%q err=%v", data, err)
	}
}

func mustNativeWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
