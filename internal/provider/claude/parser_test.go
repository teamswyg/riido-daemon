package claude

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func feedStdoutAll(t *testing.T, p agentbridge.Parser, chunks ...string) []agentbridge.RawEvent {
	t.Helper()
	var raws []agentbridge.RawEvent
	for _, chunk := range chunks {
		r, err := p.FeedStdout([]byte(chunk))
		if err != nil {
			t.Fatalf("FeedStdout %q: %v", chunk, err)
		}
		raws = append(raws, r...)
	}
	closed, err := p.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
	raws = append(raws, closed...)
	return raws
}

// One complete stream-json line per chunk → one RawEvent per chunk.
func TestParserOneLineOneEvent(t *testing.T) {
	p := NewParser()
	raws := feedStdoutAll(t, p, `{"type":"system","subtype":"init","session_id":"sess-1"}`+"\n")
	if len(raws) != 1 {
		t.Fatalf("want 1 raw event, got %d", len(raws))
	}
	if raws[0].Type != "system" {
		t.Fatalf("type: %q", raws[0].Type)
	}
	if raws[0].Source != agentbridge.RawSourceStdout {
		t.Fatalf("source: %q", raws[0].Source)
	}
	if raws[0].Payload["session_id"] != "sess-1" {
		t.Fatalf("payload missing session_id: %+v", raws[0].Payload)
	}
}

// Two lines arriving in one chunk → two RawEvents.
func TestParserMultipleLinesInOneChunk(t *testing.T) {
	p := NewParser()
	chunk := `{"type":"system"}` + "\n" + `{"type":"result"}` + "\n"
	raws := feedStdoutAll(t, p, chunk)
	if len(raws) != 2 {
		t.Fatalf("want 2 raws, got %d", len(raws))
	}
	if raws[0].Type != "system" || raws[1].Type != "result" {
		t.Fatalf("unexpected types: %q %q", raws[0].Type, raws[1].Type)
	}
}

// A line split across chunks must produce one event after the newline arrives.
func TestParserPartialLineAcrossChunks(t *testing.T) {
	p := NewParser()
	raws := feedStdoutAll(t, p,
		`{"type":"sys`,
		`tem","session_id":"x"}`+"\n",
	)
	if len(raws) != 1 {
		t.Fatalf("want 1 raw, got %d", len(raws))
	}
	if raws[0].Type != "system" || raws[0].Payload["session_id"] != "x" {
		t.Fatalf("bad reassembly: %+v", raws[0])
	}
}

// Trailing line without newline must be flushed by Close.
func TestParserUnterminatedLineFlushedOnClose(t *testing.T) {
	p := NewParser()
	raws := feedStdoutAll(t, p, `{"type":"result","subtype":"success"}`)
	if len(raws) != 1 {
		t.Fatalf("want 1 raw, got %d", len(raws))
	}
	if raws[0].Type != "result" {
		t.Fatalf("type: %q", raws[0].Type)
	}
	if raws[0].Source != agentbridge.RawSourceClose {
		t.Fatalf("trailing fragment should be tagged RawSourceClose, got %q", raws[0].Source)
	}
}

// Malformed JSON lines do NOT abort the stream — they become a RawEvent
// of Type "malformed" so the translator can decide what to do.
func TestParserMalformedJsonNotFatal(t *testing.T) {
	p := NewParser()
	chunk := "not json at all\n" + `{"type":"result"}` + "\n"
	raws := feedStdoutAll(t, p, chunk)
	if len(raws) != 2 {
		t.Fatalf("want 2 raws, got %d: %+v", len(raws), raws)
	}
	if raws[0].Type != "malformed" {
		t.Fatalf("want malformed first, got %q", raws[0].Type)
	}
	if raws[1].Type != "result" {
		t.Fatalf("second event lost: %+v", raws[1])
	}
}

// Stderr is parsed too but tagged as RawSourceStderr. Lines arriving
// with "stderr:" / "stdout:" prefix from wrapper scripts are stripped.
func TestParserStderrAndPrefixNormalization(t *testing.T) {
	p := NewParser()
	raws, err := p.FeedStderr([]byte("stderr: warning thing\n"))
	if err != nil {
		t.Fatalf("FeedStderr: %v", err)
	}
	closed, _ := p.Close()
	raws = append(raws, closed...)
	if len(raws) != 1 {
		t.Fatalf("want 1 stderr raw, got %d", len(raws))
	}
	if raws[0].Source != agentbridge.RawSourceStderr {
		t.Fatalf("source: %q", raws[0].Source)
	}
	if raws[0].Type != "stderr" {
		t.Fatalf("type: %q", raws[0].Type)
	}
	if got := string(raws[0].Bytes); got != "warning thing" {
		t.Fatalf("prefix not normalized: %q", got)
	}
}

// Lines up to ~10MB must be accepted (one big tool result).
func TestParserAcceptsLargeLine(t *testing.T) {
	p := NewParser()
	big := strings.Repeat("a", 9*1024*1024)
	chunk := `{"type":"tool_result","content":"` + big + `"}` + "\n"
	raws, err := p.FeedStdout([]byte(chunk))
	if err != nil {
		t.Fatalf("FeedStdout big line: %v", err)
	}
	if len(raws) != 1 || raws[0].Type != "tool_result" {
		t.Fatalf("big line: %+v", raws)
	}
}
