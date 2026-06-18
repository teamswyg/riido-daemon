package agentbridge

import "testing"

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
