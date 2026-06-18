package openclaw

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func goldenTranslatedEvents(raws []agentbridge.RawEvent) []agentbridge.Event {
	var out []agentbridge.Event
	for _, raw := range raws {
		evs, _, _ := Translate(raw)
		out = append(out, evs...)
	}
	return out
}
