package openclaw

import "testing"

func TestParserPrettyFullJSONResult(t *testing.T) {
	p := NewParser()

	r, err := p.FeedStdout([]byte("{\n  \"payloads\": [\n    {\"text\": \"ok\"}\n  ],\n  \"meta\": {\"agentMeta\": {\"sessionId\": \"sess-1\"}}\n}\n"))
	if err != nil {
		t.Fatalf("Feed: %v", err)
	}
	closed, _ := p.Close()
	r = append(r, closed...)

	if len(r) != 1 {
		t.Fatalf("want 1 raw, got %d: %+v", len(r), r)
	}
	if r[0].Type != "full_result" {
		t.Fatalf("type: %q", r[0].Type)
	}
}

func TestParserNDJSONFallback(t *testing.T) {
	p := NewParser()
	chunk := `{"event":"text","text":"chunk1"}` + "\n" +
		`{"event":"text","text":"chunk2"}` + "\n"

	r, _ := p.FeedStdout([]byte(chunk))
	closed, _ := p.Close()
	r = append(r, closed...)

	if len(r) != 2 {
		t.Fatalf("want 2 raws, got %d", len(r))
	}
	if r[0].Type != "ndjson:text" || r[1].Type != "ndjson:text" {
		t.Fatalf("types: %q %q", r[0].Type, r[1].Type)
	}
}
