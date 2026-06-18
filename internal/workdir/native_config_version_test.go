package workdir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeNativeConfigVersionIsDeterministicAndPolicyBound(t *testing.T) {
	_, ws := preparedTestWorkspace(t, "run-1")
	if err := NewFSAdapter("").InjectRuntimeConfig(ws, RuntimeConfig{Provider: "codex", Identity: "Agent: tester"}); err != nil {
		t.Fatal(err)
	}
	input := NativeConfigVersionInput{
		PolicyBundleVersion: "policy-bundle.test.v1",
		ProviderKind:        "codex",
		ProtocolKind:        "codex-app-server",
	}
	first, err := ComputeNativeConfigVersion(ws, input)
	if err != nil {
		t.Fatalf("ComputeNativeConfigVersion: %v", err)
	}
	assertVersionStable(t, ws, input, first)
	assertVersionChangesForPolicy(t, ws, input, first)
	assertVersionChangesForContent(t, ws, input, first)
}

func assertVersionStable(t *testing.T, ws Workspace, input NativeConfigVersionInput, first string) {
	t.Helper()
	second, err := ComputeNativeConfigVersion(ws, input)
	if err != nil || first == "" || first != second {
		t.Fatalf("version should be deterministic: first=%q second=%q err=%v", first, second, err)
	}
}

func assertVersionChangesForPolicy(t *testing.T, ws Workspace, input NativeConfigVersionInput, first string) {
	t.Helper()
	input.PolicyBundleVersion = "policy-bundle.test.v2"
	got, err := ComputeNativeConfigVersion(ws, input)
	if err != nil || got == first {
		t.Fatalf("version must change for policy: got=%q first=%q err=%v", got, first, err)
	}
}

func assertVersionChangesForContent(t *testing.T, ws Workspace, input NativeConfigVersionInput, first string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(ws.NativeConfig, "AGENTS.md"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ComputeNativeConfigVersion(ws, input)
	if err != nil || got == first {
		t.Fatalf("version must change for content: got=%q first=%q err=%v", got, first, err)
	}
}
