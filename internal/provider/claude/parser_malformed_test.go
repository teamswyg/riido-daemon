package claude

import "testing"

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
