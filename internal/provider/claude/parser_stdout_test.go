package claude

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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

func TestParserPartialLineAcrossChunks(t *testing.T) {
	p := NewParser()
	raws := feedStdoutAll(
		t, p,
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
