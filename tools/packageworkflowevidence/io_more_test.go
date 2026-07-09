package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadManifestRejectsUnknownAndTrailingJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	valid := `{"schema_version":"riido-daemon-package-workflow-evidence.v1","id":"m","loop_source":"loop","workflows":[{"id":"w","workflow":"w.yml","evidence_artifact":"e","required_fragments":["go test"]}]}`
	cases := []struct {
		name string
		body string
		want string
	}{
		{name: "unknown", body: strings.Replace(valid, `"id":"m"`, `"id":"m","extra":true`, 1), want: "unknown field"},
		{name: "trailing", body: valid + `{"schema_version":"extra"}`, want: "trailing JSON value"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustWrite(t, path, tc.body)
			_, err := readManifest(path)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}

func TestWriteJSONCreatesEvidenceFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "evidence.json")
	value := evidence{ID: "workflow", Status: "verified"}
	if err := writeJSON(path, value); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Fatalf("evidence should end with newline: %q", data)
	}
	var decoded evidence
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID != "workflow" || decoded.Status != "verified" {
		t.Fatalf("decoded evidence=%#v", decoded)
	}
}
