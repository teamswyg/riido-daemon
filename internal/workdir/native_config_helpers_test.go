package workdir

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestProviderConfigFilenameRegistry(t *testing.T) {
	for _, tc := range []struct {
		provider string
		want     string
	}{
		{"claude", "CLAUDE.md"},
		{"codex", "AGENTS.md"},
		{"openclaw", "AGENTS.md"},
		{"cursor", "AGENTS.md"},
	} {
		got := ProviderConfigFilename(tc.provider)
		if got != tc.want {
			t.Fatalf("%s: want %q, got %q", tc.provider, tc.want, got)
		}
		plan := ProviderConfigPlan(tc.provider)
		if plan.PrimaryInstructionFile != tc.want ||
			plan.ManifestFile != NativeConfigManifestPath ||
			plan.HookMode == "" {
			t.Fatalf("%s plan = %+v", tc.provider, plan)
		}
	}
	if got := ProviderConfigFilename("unknown"); got != "AGENTS.md" {
		t.Fatalf("unknown provider should fall back to AGENTS.md, got %q", got)
	}
}

func TestInjectRefusesPathTraversal(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, _ := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-Z"})
	// A malicious provider name with path traversal must NOT escape the workdir.
	err := a.InjectRuntimeConfig(ws, RuntimeConfig{Provider: "../etc"})
	if err == nil {
		t.Fatalf("expected error for path-traversal provider")
	}
}

func readNativeConfigManifest(t *testing.T, path string) NativeConfigManifest {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read native config manifest: %v", err)
	}
	var manifest NativeConfigManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatalf("decode native config manifest: %v", err)
	}
	return manifest
}

func containsString(values []string, want string) bool {
	return slices.Contains(values, want)
}
