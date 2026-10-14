package cursor

import "testing"

func TestParserStreamJSON(t *testing.T) {
	p := NewParser()
	chunk := `{"type":"system","subtype":"init","session_id":"sess-1"}` + "\n"
	r, _ := p.FeedStdout([]byte(chunk))
	if len(r) != 1 || r[0].Type != "system" {
		t.Fatalf("system: %+v", r)
	}
}

func TestParserStripsStdoutStderrPrefixes(t *testing.T) {
	p := NewParser()
	chunk := `stdout: {"type":"text","text":"hi"}` + "\n"
	r, _ := p.FeedStdout([]byte(chunk))
	if len(r) != 1 || r[0].Type != "text" {
		t.Fatalf("stdout prefix not stripped: %+v", r)
	}
}
