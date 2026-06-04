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

func TestTelemetryParserExtractsStructuredProgressCode(t *testing.T) {
	parser := NewTelemetryParser()
	events := parser.Feed(`<riido_log>{"code":1101,"args":{"label":"팀 프로젝트","description":"팀의 프로젝트 목록, 진행 상태, 우선순위와 담당자 정보를 조회해 요약을 준비 중. . ." }}<end>`)
	want := "팀 프로젝트 수집 중 - 팀의 프로젝트 목록, 진행 상태, 우선순위와 담당자 정보를 조회해 요약을 준비 중. . ."
	if len(events) != 1 ||
		events[0].Kind != EventProgress ||
		events[0].Text != want ||
		events[0].ProgressCode != 1101 ||
		events[0].ProgressKey != "tool.collecting" ||
		events[0].ProgressArgs["label"] != "팀 프로젝트" {
		t.Fatalf("events = %+v", events)
	}
}

func TestTelemetryParserRendersNumericProgressArgs(t *testing.T) {
	parser := NewTelemetryParser()
	events := parser.Feed(`<riido_log>{"code":1102,"args":{"label":"팀 프로젝트","count":3,"representative_title":"프로젝트 Alpha"}}<end>`)
	want := "팀 프로젝트 조회 완료 - 3건(프로젝트 Alpha 외)의 요약을 가져왔습니다. . ."
	if len(events) != 1 ||
		events[0].Kind != EventProgress ||
		events[0].Text != want ||
		events[0].ProgressCode != 1102 ||
		events[0].ProgressKey != "tool.collection_completed_count" ||
		events[0].ProgressArgs["count"] != "3" {
		t.Fatalf("events = %+v", events)
	}
}

func TestTelemetryParserMapsKnownLegacyProgressText(t *testing.T) {
	parser := NewTelemetryParser()
	events := parser.Feed(`<riido_log>웹 검색 실행 중 - 관련 문서(최대 3건)를 조회하고 참고 자료를 수집 중. . .<end>`)
	if len(events) != 1 ||
		events[0].ProgressCode != 1103 ||
		events[0].ProgressKey != "tool.running" ||
		events[0].ProgressArgs["label"] != "웹 검색" {
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

func TestApplyRuntimeInstructionContractPlacesByProvider(t *testing.T) {
	tests := []struct {
		provider        string
		wantInstruction string
		wantTelemetry   string
		wantPromptHas   []string
		wantSystemHas   []string
		wantPromptExact bool
	}{
		{
			provider:        "claude",
			wantInstruction: TelemetryPlacementSystemPrompt,
			wantTelemetry:   TelemetryPlacementSystemPrompt,
			wantPromptHas:   []string{"do it"},
			wantSystemHas:   []string{"Riido agent instruction:", "act as a PM", "<riido_log>"},
			wantPromptExact: true,
		},
		{
			provider:        "openclaw",
			wantInstruction: TelemetryPlacementSystemPromptInline,
			wantTelemetry:   TelemetryPlacementSystemPromptInline,
			wantPromptHas:   []string{"do it"},
			wantSystemHas:   []string{"Riido agent instruction:", "act as a PM", "<riido_log>"},
			wantPromptExact: true,
		},
		{
			provider:        "codex",
			wantInstruction: TelemetryPlacementPrompt,
			wantTelemetry:   TelemetryPlacementPrompt,
			wantPromptHas:   []string{"Riido agent instruction:", "act as a PM", "<riido_log>", "User task:", "do it"},
		},
		{
			provider:        "cursor",
			wantInstruction: TelemetryPlacementPrompt,
			wantTelemetry:   TelemetryPlacementPrompt,
			wantPromptHas:   []string{"Riido agent instruction:", "act as a PM", "<riido_log>", "User task:", "do it"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			prompt, system, telemetryPlacement, instructionPlacement := ApplyRuntimeInstructionContract(tt.provider, "do it", "", "act as a PM")
			if telemetryPlacement != tt.wantTelemetry || instructionPlacement != tt.wantInstruction {
				t.Fatalf("placements telemetry=%q instruction=%q", telemetryPlacement, instructionPlacement)
			}
			if tt.wantPromptExact && prompt != "do it" {
				t.Fatalf("prompt = %q", prompt)
			}
			for _, want := range tt.wantPromptHas {
				if !strings.Contains(prompt, want) {
					t.Fatalf("prompt missing %q: %q", want, prompt)
				}
			}
			for _, want := range tt.wantSystemHas {
				if !strings.Contains(system, want) {
					t.Fatalf("system missing %q: %q", want, system)
				}
			}
		})
	}
}

func TestRuntimeInstructionStrategiesAreStable(t *testing.T) {
	want := []RuntimeInstructionStrategy{
		{
			Provider:                  "claude",
			AgentInstructionPlacement: TelemetryPlacementSystemPrompt,
			TelemetryPlacement:        TelemetryPlacementSystemPrompt,
			EffectivenessGate:         "opt-in-real-provider-probe",
		},
		{
			Provider:                  "openclaw",
			AgentInstructionPlacement: TelemetryPlacementSystemPromptInline,
			TelemetryPlacement:        TelemetryPlacementSystemPromptInline,
			EffectivenessGate:         "opt-in-real-provider-probe",
		},
		{
			Provider:                  "codex",
			AgentInstructionPlacement: TelemetryPlacementPrompt,
			TelemetryPlacement:        TelemetryPlacementPrompt,
			EffectivenessGate:         "opt-in-real-provider-probe",
		},
		{
			Provider:                  "cursor",
			AgentInstructionPlacement: TelemetryPlacementPrompt,
			TelemetryPlacement:        TelemetryPlacementPrompt,
			EffectivenessGate:         "opt-in-real-provider-probe",
		},
	}
	got := RuntimeInstructionStrategies()
	if len(got) != len(want) {
		t.Fatalf("strategy count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("strategy[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestBuildInstructionEffectivenessProbePlacesMarkerByProvider(t *testing.T) {
	for _, strategy := range RuntimeInstructionStrategies() {
		t.Run(strategy.Provider, func(t *testing.T) {
			probe := BuildInstructionEffectivenessProbe(strategy.Provider)
			if probe.AgentInstructionPlacement != strategy.AgentInstructionPlacement {
				t.Fatalf("instruction placement = %q, want %q", probe.AgentInstructionPlacement, strategy.AgentInstructionPlacement)
			}
			if probe.TelemetryPlacement != strategy.TelemetryPlacement {
				t.Fatalf("telemetry placement = %q, want %q", probe.TelemetryPlacement, strategy.TelemetryPlacement)
			}
			promptHasMarker := strings.Contains(probe.Prompt, probe.ExpectedMarker)
			systemHasMarker := strings.Contains(probe.SystemPrompt, probe.ExpectedMarker)
			switch strategy.AgentInstructionPlacement {
			case TelemetryPlacementPrompt:
				if !promptHasMarker || systemHasMarker {
					t.Fatalf("marker location prompt=%t system=%t probe=%+v", promptHasMarker, systemHasMarker, probe)
				}
			case TelemetryPlacementSystemPrompt, TelemetryPlacementSystemPromptInline:
				if promptHasMarker || !systemHasMarker {
					t.Fatalf("marker location prompt=%t system=%t probe=%+v", promptHasMarker, systemHasMarker, probe)
				}
			default:
				t.Fatalf("unknown placement %q", strategy.AgentInstructionPlacement)
			}
			if err := ValidateInstructionEffectivenessOutput(probe, "ok\n"+probe.ExpectedMarker+"\n"); err != nil {
				t.Fatalf("valid marker rejected: %v", err)
			}
			if err := ValidateInstructionEffectivenessOutput(probe, "ok"); err == nil {
				t.Fatal("missing marker accepted")
			}
		})
	}
}

func TestRuntimeInstructionStrategyForUnknownProviderDefaultsToPrompt(t *testing.T) {
	strategy := RuntimeInstructionStrategyForProvider("future-provider")
	if strategy.Provider != "future-provider" {
		t.Fatalf("provider = %q", strategy.Provider)
	}
	if strategy.AgentInstructionPlacement != TelemetryPlacementPrompt || strategy.TelemetryPlacement != TelemetryPlacementPrompt {
		t.Fatalf("unknown provider strategy = %+v", strategy)
	}
}

func TestInjectTelemetryContractIsIdempotent(t *testing.T) {
	first := InjectTelemetryContract("do it")
	second := InjectTelemetryContract(first)
	if second != first {
		t.Fatalf("telemetry contract duplicated:\nfirst=%q\nsecond=%q", first, second)
	}
}
