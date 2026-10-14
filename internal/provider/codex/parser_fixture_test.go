package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func feedAll(t *testing.T, p agentbridge.Parser, chunks ...string) []agentbridge.RawEvent {
	t.Helper()
	var raws []agentbridge.RawEvent
	for _, c := range chunks {
		r, err := p.FeedStdout([]byte(c))
		if err != nil {
			t.Fatalf("FeedStdout %q: %v", c, err)
		}
		raws = append(raws, r...)
	}
	closed, err := p.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
	return append(raws, closed...)
}
