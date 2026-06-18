package cursor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func runFixtureThroughParser(t *testing.T, name string) []agentbridge.Event {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	p := NewParser()
	raws, err := p.FeedStdout(body)
	if err != nil {
		t.Fatalf("FeedStdout: %v", err)
	}
	closed, _ := p.Close()
	return translateRaws(t, append(raws, closed...))
}

func translateRaws(t *testing.T, raws []agentbridge.RawEvent) []agentbridge.Event {
	t.Helper()
	var events []agentbridge.Event
	for _, raw := range raws {
		evs, _, err := Translate(raw)
		if err != nil {
			t.Fatalf("translate: %v", err)
		}
		events = append(events, evs...)
	}
	return events
}
