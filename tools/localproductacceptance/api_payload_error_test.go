package main

import "testing"

func TestPayloadErrorSummaryUsesKnownErrorFields(t *testing.T) {
	got := payloadErrorSummary(map[string]any{"error": "runtime binding missing"})
	if got != "runtime binding missing" {
		t.Fatalf("summary=%q", got)
	}
}

func TestAPIRepairForPayloadClassifiesRuntimeBinding(t *testing.T) {
	got := apiRepairForPayload(map[string]any{"error": "ai agent runtime binding is not configured"})
	if got == nil || got.Class != "ai_agent_runtime_binding_required" {
		t.Fatalf("repair=%+v", got)
	}
}
