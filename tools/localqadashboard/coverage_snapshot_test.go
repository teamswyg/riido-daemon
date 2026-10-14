package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteCoverageSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "coverage.json")
	rows := []coverageRow{{
		ID: "provider.cursor", Title: "Cursor", Status: "skipped",
		Evidence: "provider.json", ExpiresAt: "2999-01-01T00:00:00Z",
	}}
	summary := coverageSummary{Total: 1, Skipped: 1}
	if err := writeCoverageSnapshot(path, rows, summary); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got coverageSnapshot
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Summary.Total != 1 || got.Rows[0].ID != "provider.cursor" {
		t.Fatalf("snapshot=%+v", got)
	}
	if got.Rows[0].Repair != nil {
		t.Fatalf("snapshot=%+v", got)
	}
	if got.Rows[0].Evidence != "provider.json" || got.Rows[0].ExpiresAt == "" {
		t.Fatalf("snapshot provenance missing: %+v", got.Rows[0])
	}
}
