package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHydrateConfigFromStorageKeepsExplicitValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	data := `{"cookies":[{"name":"token","value":"from-storage"}],"origins":[{"localStorage":[{"name":"workspace-from-storage-toggle"}]}]}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	token, workspace := "explicit-token", "explicit-workspace"
	cfg := storageTestConfig(path, &token, &workspace)

	hydrateConfigFromStorage(cfg)

	if token != "explicit-token" || workspace != "explicit-workspace" {
		t.Fatalf("explicit config was overwritten: token=%q workspace=%q", token, workspace)
	}
}

func TestLoadStorageStateRejectsMissingBlankAndInvalid(t *testing.T) {
	if _, ok := loadStorageState(""); ok {
		t.Fatal("blank storage state path should be ignored")
	}
	if _, ok := loadStorageState(filepath.Join(t.TempDir(), "missing.json")); ok {
		t.Fatal("missing storage state path should be ignored")
	}
	path := filepath.Join(t.TempDir(), "broken.json")
	if err := os.WriteFile(path, []byte(`{`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, ok := loadStorageState(path); ok {
		t.Fatal("invalid storage state should be ignored")
	}
}

func TestStorageStateExtractorsAreConservative(t *testing.T) {
	state := storageState{
		Cookies: []storageCookie{
			{Name: "other", Value: "x"},
			{Name: "token", Value: "  tok  "},
		},
		Origins: []storageOrigin{{LocalStorage: []storageEntry{
			{Name: "workspace--toggle"},
			{Name: "workspace-ws-a-toggle"},
		}}},
	}
	if got := stateToken(state); got != "tok" {
		t.Fatalf("token=%q", got)
	}
	if got := stateWorkspaceID(state); got != "ws-a" {
		t.Fatalf("workspace=%q", got)
	}
	if id, ok := parseWorkspaceKey("workspace-ws-a"); ok || id != "" {
		t.Fatalf("invalid key parsed as id=%q ok=%v", id, ok)
	}
}
