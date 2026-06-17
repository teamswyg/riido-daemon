package agentexecutionevidence

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"
)

func loadManifest(t *testing.T, path string) evidenceManifest {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open manifest: %v", err)
	}
	defer file.Close()

	var manifest evidenceManifest
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		t.Fatalf("manifest must contain a single JSON object, trailing decode: %v", err)
	}
	return manifest
}

func readText(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
