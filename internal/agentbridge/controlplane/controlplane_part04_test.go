package controlplane

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readFileRuntimeRecord(t *testing.T, path string) RegisteredRuntime {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read runtime record: %v", err)
	}
	var rec RegisteredRuntime
	if err := json.Unmarshal(body, &rec); err != nil {
		t.Fatalf("decode runtime record: %v", err)
	}
	return rec
}

func countTopLevelJSON(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".json" {
			count++
		}
	}
	return count
}

func readClaimRecords(t *testing.T, dir string) []FileClaimRecord {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read claims dir: %v", err)
	}
	records := make([]FileClaimRecord, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var rec FileClaimRecord
		if err := json.Unmarshal(body, &rec); err != nil {
			t.Fatalf("decode claim record %s: %v", entry.Name(), err)
		}
		records = append(records, rec)
	}
	return records
}
