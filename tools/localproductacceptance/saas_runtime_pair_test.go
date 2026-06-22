package main

import "testing"

func TestChoosePreparedRuntimePairPrefersSameCodexKind(t *testing.T) {
	runtimes := []preparedRuntime{
		{RuntimeID: "dev-a:openclaw", Kind: "openclaw"},
		{RuntimeID: "dev-a:codex", Kind: "codex"},
		{RuntimeID: "dev-b:codex", Kind: "codex"},
	}
	pair, ok := choosePreparedRuntimePair(runtimes)
	if !ok || pair[0].Kind != "codex" || pair[0].RuntimeID == pair[1].RuntimeID {
		t.Fatalf("pair=%+v ok=%v", pair, ok)
	}
}

func TestPreparedRuntimesReadyRequiresProviderVersion(t *testing.T) {
	payload := map[string]any{"devices": []any{
		map[string]any{"device_id": "dev-a", "runtimes": []any{runtimePayload("dev-a:codex", "codex", "")}},
		map[string]any{"device_id": "dev-b", "runtimes": []any{runtimePayload("dev-b:codex", "codex", "codex live")}},
	}}
	if preparedRuntimesReady(payload, []string{"dev-a", "dev-b"}) {
		t.Fatal("runtime pair without both provider versions must not be ready")
	}
}

func runtimePayload(runtimeID, kind, version string) map[string]any {
	return map[string]any{
		"runtime_id": runtimeID, "kind": kind, "provider_version": version,
		"availability": "online", "detection_state": "detected",
	}
}
