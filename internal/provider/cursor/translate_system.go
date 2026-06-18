package cursor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func translateSystem(p map[string]any) []agentbridge.Event {
	var out []agentbridge.Event
	if sid := stringField(p, "session_id"); sid != "" {
		out = append(out, agentbridge.Event{Kind: agentbridge.EventSessionIdentified, SessionID: sid})
	}
	return append(out, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning})
}
