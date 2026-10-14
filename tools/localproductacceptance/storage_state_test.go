package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHydrateConfigFromStorage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	data := `{"cookies":[{"name":"token","value":"tok"}],"origins":[{"localStorage":[{"name":"workspace-ws-1-toggle"}]}]}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	token, workspace := "", ""
	cfg := storageTestConfig(path, &token, &workspace)

	hydrateConfigFromStorage(cfg)

	if token != "tok" || workspace != "ws-1" {
		t.Fatalf("token=%q workspace=%q", token, workspace)
	}
}

func storageTestConfig(path string, token, workspace *string) config {
	return config{
		apiToken:     token,
		workspaceID:  workspace,
		storageState: &path,
	}
}
