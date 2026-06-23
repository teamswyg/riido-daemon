package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteCoverageSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "coverage.json")
	rows := []coverageRow{{ID: "provider.cursor", Title: "Cursor", Status: "skipped"}}
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
}
