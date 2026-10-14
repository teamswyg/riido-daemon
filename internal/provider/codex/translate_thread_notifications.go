package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func codexThreadStartedEvent(sessionID string) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind:      agentbridge.EventSessionIdentified,
		SessionID: sessionID,
	}}
}

func threadIDFromParams(p map[string]any) string {
	if id := stringField(p, "threadId"); id != "" {
		return id
	}
	if id := stringField(p, "thread_id"); id != "" {
		return id
	}
	thread := mapField(p, "thread")
	if id := stringField(thread, "id"); id != "" {
		return id
	}
	return stringField(thread, "sessionId")
}
