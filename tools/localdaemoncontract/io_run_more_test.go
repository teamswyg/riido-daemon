package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestRejectsUnknownAndTrailingJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	valid := `{"schema_version":"v","id":"x","title":"X","generated_doc":"doc.md","workflow":"wf.yml","evidence_artifact":"e","purpose":"p","facts":[{"name":"f","summary":"s","source_checks":["src"]}],"boundaries":[],"source_checks":[{"name":"src","file":"src.go","contains":"needle"}],"assertions":[]}`
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
			mustLocalDaemonWrite(t, path, tc.body)
			_, err := loadManifest(dir, "manifest.json")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}

func TestRunWritesDocAndEvidence(t *testing.T) {
	dir := t.TempDir()
	manifest := localDaemonManifest()
	mustLocalDaemonWrite(t, filepath.Join(dir, "src.go"), "needle")
	writeLocalDaemonManifest(t, dir, manifest)
	out := filepath.Join(dir, "out", "evidence.json")
	if err := run(options{Repo: dir, Manifest: "manifest.json", WriteDoc: true, CheckDoc: true, EvidenceOut: out}); err != nil {
		t.Fatal(err)
	}
	doc, err := os.ReadFile(filepath.Join(dir, manifest.GeneratedDoc))
	if err != nil || !strings.Contains(string(doc), manifest.Title) {
		t.Fatalf("doc=%q err=%v", doc, err)
	}
	var evidence Evidence
	data, err := os.ReadFile(out)
	if err != nil || json.Unmarshal(data, &evidence) != nil {
		t.Fatalf("evidence read/unmarshal failed: %v %s", err, data)
	}
	if evidence.ID != manifest.ID || len(evidence.SourceChecks) != 1 {
		t.Fatalf("evidence=%#v", evidence)
	}
}

func writeLocalDaemonManifest(t *testing.T, dir string, m Manifest) {
	t.Helper()
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	mustLocalDaemonWrite(t, filepath.Join(dir, "manifest.json"), string(data))
}
