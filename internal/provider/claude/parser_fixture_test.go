package claude

import (
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
	return append(raws, closed...)
}
