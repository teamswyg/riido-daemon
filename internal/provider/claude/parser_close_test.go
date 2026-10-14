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
