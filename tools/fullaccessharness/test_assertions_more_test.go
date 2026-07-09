package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func mustFullAccessJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertFullAccessError(t *testing.T, err error, needle string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), needle) {
		t.Fatalf("expected %q error, got %v", needle, err)
	}
}
