package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentManifestAndGeneratedDocs(t *testing.T) {
	if err := run("../..", "docs/readme/document-map.riido.json", "", false, true); err != nil {
		t.Fatal(err)
	}
}

func TestRejectsMissingMappedDoc(t *testing.T) {
	dir, manifestPath := fixture(t)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	mutated := strings.Replace(string(data), "docs/a.md", "docs/missing.md", 1)
	mustWrite(t, manifestPath, mutated)
	err = run(dir, manifestPath, "", false, false)
	if err == nil || !strings.Contains(err.Error(), `missing doc "docs/missing.md"`) {
		t.Fatalf("expected missing doc error, got %v", err)
	}
}

func TestEvidenceOutputSummarizesMap(t *testing.T) {
	dir, manifestPath := fixture(t)
	evidencePath := filepath.Join(dir, "evidence.json")
	if err := run(dir, manifestPath, evidencePath, true, true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatal(err)
	}
	var evidence evidenceFile
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.Status != "verified" || evidence.DecisionCount != 1 || evidence.ReadOrderCount != 1 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}
