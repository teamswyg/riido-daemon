package supervisor

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
)

func readRunEvents(t *testing.T, path string) []ir.CanonicalEvent {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read run event log: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	out := make([]ir.CanonicalEvent, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var ev ir.CanonicalEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("decode run event %q: %v", line, err)
		}
		out = append(out, ev)
	}

	return out
}

func assertRunEvent(
	t *testing.T,
	events []ir.CanonicalEvent,
	eventType ir.EventType,
	check func(ir.CanonicalEvent),
) {
	t.Helper()

	for _, ev := range events {
		if ev.Type == eventType {
			if check != nil {
				check(ev)
			}
			return
		}
	}

	t.Fatalf("run event %s not found in %+v", eventType, events)
}
