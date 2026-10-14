package openclaw

import "testing"

func TestParserFullJSONResult(t *testing.T) {
	p := NewParser()
	r, err := p.FeedStdout([]byte(`{"session_id":"sess-1","text":"hello","usage":{"prompt_tokens":3,"completion_tokens":7}}`))
	if err != nil {
		t.Fatalf("Feed: %v", err)
	}
	closed, _ := p.Close()
	r = append(r, closed...)
	if len(r) != 1 {
		t.Fatalf("want 1 raw, got %d", len(r))
	}
	if r[0].Type != "full_result" {
		t.Fatalf("type: %q", r[0].Type)
	}
}
