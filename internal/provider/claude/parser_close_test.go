package claude

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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
		t.Fatalf("trailing fragment should be RawSourceClose, got %q", raws[0].Source)
	}
}

func TestParserUnterminatedStderrKeepsSourceOnClose(t *testing.T) {
	p := NewParser()
	_, _ = p.FeedStderr([]byte("stderr: invalid launch configuration"))
	raws, err := p.Close()
	if err != nil || len(raws) != 1 || raws[0].Source != agentbridge.RawSourceStderr {
		t.Fatalf("trailing stderr lost provenance: %v %v", raws, err)
	}
	events, _, err := Translate(raws[0])
	if err != nil || len(events) != 1 || events[0].Text != "invalid launch configuration" {
		t.Fatalf("trailing stderr lost diagnostic: %v %v", events, err)
	}
}
