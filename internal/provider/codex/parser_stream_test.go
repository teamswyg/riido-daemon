package codex

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestParserPartialLineReassembly(t *testing.T) {
	p := NewParser()
	raws := feedAll(t, p, `{"jsonrpc":"2.0","method":"x`, `"}`+"\n")
	if len(raws) != 1 {
		t.Fatalf("reassembly: %+v", raws)
	}
}

func TestParserMalformedNonFatal(t *testing.T) {
	chunk := "garbage\n" + `{"jsonrpc":"2.0","method":"ok"}` + "\n"
	raws := feedAll(t, NewParser(), chunk)
	if len(raws) != 2 {
		t.Fatalf("want 2, got %d", len(raws))
	}
	if raws[0].Type != "malformed" {
		t.Fatalf("first must be malformed: %+v", raws[0])
	}
	if raws[1].Type != "notification:ok" {
		t.Fatalf("second must be ok: %+v", raws[1])
	}
}

func TestParserStderrTagged(t *testing.T) {
	p := NewParser()
	r, _ := p.FeedStderr([]byte("warn line\n"))
	if len(r) != 1 {
		t.Fatalf("want 1, got %d", len(r))
	}
	if r[0].Source != agentbridge.RawSourceStderr || r[0].Type != "stderr" {
		t.Fatalf("source/type: %+v", r[0])
	}
	if !strings.Contains(string(r[0].Bytes), "warn line") {
		t.Fatalf("bytes: %q", r[0].Bytes)
	}
}
