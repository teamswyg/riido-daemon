package claude

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func translateSystem(raw agentbridge.RawEvent) []agentbridge.Event {
	var out []agentbridge.Event
	if sid := stringField(raw.Payload, "session_id"); sid != "" {
		out = append(out, agentbridge.Event{Kind: agentbridge.EventSessionIdentified, SessionID: sid})
	}
	out = append(out, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning})
	return out
}
