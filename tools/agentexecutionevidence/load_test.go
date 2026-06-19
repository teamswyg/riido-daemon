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
	return loadJSONFile[evidenceManifest](t, path)
}

func loadJSONFile[T any](t *testing.T, path string) T {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer file.Close()

	var value T
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		t.Fatalf("%s must contain a single JSON value, trailing decode: %v", path, err)
	}
	return value
}

func readText(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
