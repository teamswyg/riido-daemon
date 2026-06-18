package taskdbplane

import (
	"encoding/json"
	"os"
	"testing"
)

func readRuntimeRegistry(t *testing.T, path string) RuntimeRegistry {
	t.Helper()
	var registry RuntimeRegistry
	readJSONFixture(t, path, &registry)
	return registry
}

func readRuntimeLeaseRegistry(t *testing.T, path string) RuntimeLeaseRegistry {
	t.Helper()
	var registry RuntimeLeaseRegistry
	readJSONFixture(t, path, &registry)
	return registry
}

func readJSONFixture(t *testing.T, path string, out any) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	if err := json.Unmarshal(body, out); err != nil {
		t.Fatalf("decode registry: %v", err)
	}
}
