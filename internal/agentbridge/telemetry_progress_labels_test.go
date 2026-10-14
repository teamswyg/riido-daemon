package agentbridge

import "testing"

func TestTelemetryParserNormalizesStatefulStructuredProgressLabels(t *testing.T) {
	parser := NewTelemetryParser()
	events := parser.Feed(
		`<riido_log>{"code":1103,"args":{"label":"테스트 실행","description":"Rust 프로젝트에서 cargo test를 다시 실행합니다."}}<end>` +
			`<riido_log>{"code":1104,"args":{"label":"검증 완료","summary":"README 마지막 줄과 cargo test 통과 결과를 확인했습니다."}}<end>`,
	)
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
