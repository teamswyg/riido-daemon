package agentbridge

import (
	"strings"
	"testing"
)

func TestTelemetryParserExtractsRiidoLog(t *testing.T) {
	parser := NewTelemetryParser()
	events := parser.Feed(`before <riido_log>프로젝트 go.mod 작성중<end> after`)
	if len(events) != 1 || events[0].Kind != EventProgress || events[0].Text != "프로젝트 go.mod 작성중" {
		t.Fatalf("events = %+v", events)
	}
}

func TestTelemetryParserHandlesSplitTags(t *testing.T) {
	parser := NewTelemetryParser()
	if events := parser.Feed("noise <riido_"); len(events) != 0 {
		t.Fatalf("unexpected early events: %+v", events)
	}
	if events := parser.Feed("log>main.go 작성중<e"); len(events) != 0 {
		t.Fatalf("unexpected partial end events: %+v", events)
	}
	events := parser.Feed("nd>")
	if len(events) != 1 || events[0].Text != "main.go 작성중" {
		t.Fatalf("events = %+v", events)
	}
}

func TestTelemetryParserExtractsMultipleMessages(t *testing.T) {
	parser := NewTelemetryParser()
	events := parser.Feed(`<riido_log>go.mod<end><riido_log>go test<end>`)
	if len(events) != 2 || events[0].Text != "go.mod" || events[1].Text != "go test" {
		t.Fatalf("events = %+v", events)
	}
}

func TestInjectTelemetryContract(t *testing.T) {
	prompt := InjectTelemetryContract("golang hello world 빠르게 만들어줘")
	if !strings.Contains(prompt, "<riido_log>") || !strings.Contains(prompt, "<end>") {
		t.Fatalf("telemetry contract missing tags: %q", prompt)
	}
	if !strings.Contains(prompt, "golang hello world") {
		t.Fatalf("original prompt missing: %q", prompt)
	}
}

func TestApplyTelemetryContractPlacesByProvider(t *testing.T) {
	codexPrompt, codexSystem, codexPlacement := ApplyTelemetryContract("codex", "do it", "")
	if codexPlacement != TelemetryPlacementPrompt || !strings.Contains(codexPrompt, "<riido_log>") || codexSystem != "" {
		t.Fatalf("codex placement prompt=%q system=%q placement=%q", codexPrompt, codexSystem, codexPlacement)
	}
	claudePrompt, claudeSystem, claudePlacement := ApplyTelemetryContract("claude", "do it", "be concise")
	if claudePlacement != TelemetryPlacementSystemPrompt || claudePrompt != "do it" || !strings.Contains(claudeSystem, "<riido_log>") || !strings.Contains(claudeSystem, "be concise") {
		t.Fatalf("claude placement prompt=%q system=%q placement=%q", claudePrompt, claudeSystem, claudePlacement)
	}
	openClawPrompt, openClawSystem, openClawPlacement := ApplyTelemetryContract("openclaw", "do it", "")
	if openClawPlacement != TelemetryPlacementSystemPromptInline || openClawPrompt != "do it" || !strings.Contains(openClawSystem, "<riido_log>") {
		t.Fatalf("openclaw placement prompt=%q system=%q placement=%q", openClawPrompt, openClawSystem, openClawPlacement)
	}
}

func TestInjectTelemetryContractIsIdempotent(t *testing.T) {
	first := InjectTelemetryContract("do it")
	second := InjectTelemetryContract(first)
	if second != first {
		t.Fatalf("telemetry contract duplicated:\nfirst=%q\nsecond=%q", first, second)
	}
}
