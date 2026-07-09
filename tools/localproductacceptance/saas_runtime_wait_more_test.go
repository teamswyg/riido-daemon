package main

import (
	"errors"
	"net/http"
	"testing"
)

func TestRuntimeWaitObservedRecordsPreparedPair(t *testing.T) {
	payload := map[string]any{"devices": []any{
		map[string]any{"device_id": "dev-a", "runtimes": []any{runtimePayload("rt-a", "codex", "v1")}},
		map[string]any{"device_id": "dev-b", "runtimes": []any{runtimePayload("rt-b", "codex", "v2")}},
		map[string]any{"device_id": "other", "runtimes": []any{runtimePayload("rt-c", "codex", "v3")}},
	}}
	if !preparedRuntimesReady(payload, []string{"dev-a", "dev-b"}) {
		t.Fatal("expected ready pair with provider versions")
	}
	got := runtimeWaitObserved(http.StatusOK, payload, []string{"dev-a", "dev-b"})
	if got["prepared_runtimes_count"] != 2 || got["same_runtime_kind_pair"] != true {
		t.Fatalf("unexpected observed map: %+v", got)
	}
	if got["runtime_kind"] != "codex" || got["first_runtime_id"] != "rt-a" {
		t.Fatalf("missing pair identity: %+v", got)
	}
}

func TestRuntimeWaitScenarioBuilders(t *testing.T) {
	payload := map[string]any{"devices": []any{}}
	passed := passedRuntimeWaitScenario("/v2/client/workspaces/ws/ai-agent", http.StatusOK, payload, []string{"dev-a"})
	if passed.Status != statusPassed || passed.Endpoint != "/v2/client/workspaces/ws/ai-agent/devices" {
		t.Fatalf("unexpected passed scenario: %+v", passed)
	}
	failed := failedRuntimeWaitScenario("base", 0, nil, payload, []string{"dev-a"})
	if failed.Status != statusFailed || failed.FailureSummary == "" {
		t.Fatalf("unexpected failed scenario: %+v", failed)
	}
	withErr := failedRuntimeWaitScenario("base", 0, errors.New("boom"), payload, nil)
	if withErr.FailureSummary != "boom" {
		t.Fatalf("expected error summary, got %q", withErr.FailureSummary)
	}
}

func TestSaaSPrepareHelpers(t *testing.T) {
	if normalizedDaemonSlots(0) != 2 || normalizedDaemonSlots(3) != 3 {
		t.Fatal("daemon slots should default to at least two")
	}
	if startScenarioPassed(nil) {
		t.Fatal("empty scenario list should not pass")
	}
	if startScenarioPassed([]scenario{{Status: statusPassed}}) {
		t.Fatal("build-only scenario list should not pass")
	}
	scenarios := []scenario{{Status: statusPassed}, {Status: statusPassed}}
	if !startScenarioPassed(scenarios) {
		t.Fatal("second start scenario should determine pass")
	}
	scenarios[1].Status = statusFailed
	if startScenarioPassed(scenarios) {
		t.Fatal("failed start scenario should not pass")
	}
}
