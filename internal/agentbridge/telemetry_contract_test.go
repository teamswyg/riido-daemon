package agentbridge

import (
	"strings"
	"testing"
)

func TestInjectTelemetryContract(t *testing.T) {
	prompt := InjectTelemetryContract("golang hello world 빠르게 만들어줘")
	if !strings.Contains(prompt, "<riido_log>") || !strings.Contains(prompt, "<end>") {
		t.Fatalf("telemetry contract missing tags: %q", prompt)
	}
	if !strings.Contains(prompt, "golang hello world") {
		t.Fatalf("original prompt missing: %q", prompt)
	}
	if !strings.Contains(prompt, "state-neutral") {
		t.Fatalf("state-neutral label rule missing: %q", prompt)
	}
}

func TestApplyTelemetryContractPlacesByProvider(t *testing.T) {
	codexPrompt, codexSystem, codexPlacement := ApplyTelemetryContract("codex", "do it", "")
	if codexPlacement != TelemetryPlacementPrompt ||
		!strings.Contains(codexPrompt, "<riido_log>") ||
		codexSystem != "" {
		t.Fatalf("codex placement prompt=%q system=%q placement=%q", codexPrompt, codexSystem, codexPlacement)
	}
	claudePrompt, claudeSystem, claudePlacement := ApplyTelemetryContract("claude", "do it", "be concise")
	if claudePlacement != TelemetryPlacementSystemPrompt ||
		claudePrompt != "do it" ||
		!strings.Contains(claudeSystem, "<riido_log>") ||
		!strings.Contains(claudeSystem, "be concise") {
		t.Fatalf("claude placement prompt=%q system=%q placement=%q", claudePrompt, claudeSystem, claudePlacement)
	}
	openClawPrompt, openClawSystem, openClawPlacement := ApplyTelemetryContract("openclaw", "do it", "")
	if openClawPlacement != TelemetryPlacementSystemPromptInline ||
		openClawPrompt != "do it" ||
		!strings.Contains(openClawSystem, "<riido_log>") {
		t.Fatalf("openclaw placement prompt=%q system=%q placement=%q", openClawPrompt, openClawSystem, openClawPlacement)
	}
}
