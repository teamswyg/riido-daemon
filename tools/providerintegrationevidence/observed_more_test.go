package main

import (
	"strings"
	"testing"
)

func TestProviderObservedAddsOpenClawModelHints(t *testing.T) {
	got := providerObserved(provider{ID: "openclaw"}, "")
	if got["model_override_env"] != "RIIDO_OPENCLAW_INTEGRATION_MODEL" {
		t.Fatalf("observed=%#v", got)
	}
	candidates, ok := got["integration_model_candidates"].([]string)
	if !ok || len(candidates) == 0 {
		t.Fatalf("model candidates missing: %#v", got)
	}
}

func TestProviderObservedUnknownProviderIsNil(t *testing.T) {
	if got := providerObserved(provider{ID: "claude"}, "/bin/claude"); got != nil {
		t.Fatalf("unexpected provider observed payload: %#v", got)
	}
}

func TestCompactOutputKeepsTailForLargeProviderLogs(t *testing.T) {
	long := strings.Repeat("a", 700) + "tail"
	got := compactOutput("  " + long + "  ")
	if len(got) != 600 {
		t.Fatalf("compact length=%d", len(got))
	}
	if !strings.HasSuffix(got, "tail") {
		t.Fatalf("compact output lost tail: %q", got[len(got)-16:])
	}
}

func TestJoinProblemsFormatsEvidenceBullets(t *testing.T) {
	got := joinProblems([]string{"missing claude", "cursor auth"})
	if got != "- missing claude\n- cursor auth\n" {
		t.Fatalf("problems=%q", got)
	}
}

func TestOpenClawConfigRepairRemainsManual(t *testing.T) {
	got := openClawConfigRepair()
	if got.Class != "provider_config_invalid" || got.Mode != "manual" {
		t.Fatalf("repair=%+v", got)
	}
	if !strings.Contains(got.SuggestedCommand, "openclaw doctor --fix") {
		t.Fatalf("repair command=%q", got.SuggestedCommand)
	}
}
