package agentbridge

import "testing"

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

func TestTelemetryParserFiltersSplitTelemetryFromVisibleText(t *testing.T) {
	parser := NewTelemetryParser()
	visible, events := parser.FilterTextDelta("생각 중...\n<ri")
	if visible != "생각 중...\n" || len(events) != 0 {
		t.Fatalf("first chunk visible=%q events=%+v", visible, events)
	}
	visible, events = parser.FilterTextDelta(`ido_log>{"code":1001,"args":{}}<e`)
	if visible != "" || len(events) != 0 {
		t.Fatalf("middle chunk visible=%q events=%+v", visible, events)
	}
	visible, events = parser.FilterTextDelta("nd>계속합니다")
	if visible != "계속합니다" || len(events) != 1 || events[0].Kind != EventProgress {
		t.Fatalf("final chunk visible=%q events=%+v", visible, events)
	}
}

func TestTelemetryParserExtractsMultipleMessages(t *testing.T) {
	parser := NewTelemetryParser()
	events := parser.Feed(`<riido_log>go.mod<end><riido_log>go test<end>`)
	if len(events) != 2 || events[0].Text != "go.mod" || events[1].Text != "go test" {
		t.Fatalf("events = %+v", events)
	}
}
