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
		events[0].ProgressCode != ProgressCodeToolCollecting ||
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
		events[0].ProgressCode != ProgressCodeToolCollectionCompletedCount ||
		events[0].ProgressKey != "tool.collection_completed_count" ||
		events[0].ProgressArgs["count"] != "3" {
		t.Fatalf("events = %+v", events)
	}
}

func TestTelemetryParserNormalizesStatefulStructuredProgressLabels(t *testing.T) {
	parser := NewTelemetryParser()
	events := parser.Feed(`<riido_log>{"code":1103,"args":{"label":"테스트 실행","description":"Rust 프로젝트에서 cargo test를 다시 실행합니다."}}<end><riido_log>{"code":1104,"args":{"label":"검증 완료","summary":"README 마지막 줄과 cargo test 통과 결과를 확인했습니다."}}<end>`)
	if len(events) != 2 {
		t.Fatalf("events = %+v", events)
	}
	if events[0].Text != "테스트 실행 중 - Rust 프로젝트에서 cargo test를 다시 실행합니다." ||
		events[0].ProgressCode != ProgressCodeToolRunning ||
		events[0].ProgressArgs["label"] != "테스트" {
		t.Fatalf("running event = %+v", events[0])
	}
	if events[1].Text != "검증 완료 - README 마지막 줄과 cargo test 통과 결과를 확인했습니다." ||
		events[1].ProgressCode != ProgressCodeToolCompleted ||
		events[1].ProgressArgs["label"] != "검증" {
		t.Fatalf("completed event = %+v", events[1])
	}
}

func TestTelemetryParserMapsKnownLegacyProgressText(t *testing.T) {
	parser := NewTelemetryParser()
	events := parser.Feed(`<riido_log>웹 검색 실행 중 - 관련 문서(최대 3건)를 조회하고 참고 자료를 수집 중. . .<end>`)
	if len(events) != 1 ||
		events[0].ProgressCode != ProgressCodeToolRunning ||
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
	if !strings.Contains(prompt, "state-neutral") {
		t.Fatalf("state-neutral label rule missing: %q", prompt)
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
