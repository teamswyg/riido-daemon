package claude

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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
