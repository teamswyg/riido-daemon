package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestStrictJSON(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir)
	manifestPath := filepath.Join(dir, "manifest.json")
	body, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		body string
		want string
	}{
		{
			name: "unknown field",
			body: strings.Replace(string(body), `"id":"x"`, `"id":"x","extra":true`, 1),
			want: "unknown field",
		},
		{
			name: "trailing value",
			body: string(body) + `{"schema_version":"other"}`,
			want: "trailing JSON value",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustWrite(t, manifestPath, tc.body)
			_, err := loadManifest(dir, "manifest.json")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}

func TestWritersCreateParentDirectories(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "nested", "evidence", "out.json")
	if err := writeJSON(jsonPath, Evidence{ID: "e"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"id": "e"`) {
		t.Fatalf("json evidence missing id: %s", data)
	}
	textPath := filepath.Join(dir, "nested", "docs", "out.md")
	if err := writeText(textPath, "hello"); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(textPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected text output %q", data)
	}
}
