package openclaw

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func rawBytesEvent(
	source agentbridge.RawSource,
	eventType string,
	body []byte,
) agentbridge.RawEvent {
	return agentbridge.RawEvent{
		Source: source,
		Type:   eventType,
		Bytes:  append([]byte(nil), body...),
	}
}

func rawPayloadEvent(
	source agentbridge.RawSource,
	eventType string,
	payload map[string]any,
	body []byte,
) agentbridge.RawEvent {
	ev := rawBytesEvent(source, eventType, body)
	ev.Payload = payload
	return ev
}
