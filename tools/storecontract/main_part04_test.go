package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func removeString(items []string, unwanted string) []string {
	var out []string
	for _, item := range items {
		if item != unwanted {
			out = append(out, item)
		}
	}
	return out
}

func hasError(errors []string, wanted string) bool {
	return slices.Contains(errors, wanted)
}
