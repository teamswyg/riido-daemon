package cursor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

// Translate maps a Cursor RawEvent to run-scope Events.
func Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	switch raw.Source {
	case agentbridge.RawSourceStderr:
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: string(raw.Bytes)}}, nil, nil
	case agentbridge.RawSourceStdout, agentbridge.RawSourceClose:
	}

	switch wireEventType(raw.Type) {
	case wireEventMalformed:
		return malformedEvent(raw), nil, nil
	case wireEventSystem:
		return translateSystem(raw.Payload), nil, nil
	case wireEventText:
		return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: stringField(raw.Payload, "text")}}, nil, nil
	case wireEventAssistant:
		return translateAssistant(raw.Payload), nil, nil
	case wireEventToolUse:
		return []agentbridge.Event{toolStartedFromPayload(raw.Payload)}, nil, nil
	case wireEventToolResult:
		return []agentbridge.Event{toolResultFromPayload(raw.Payload)}, nil, nil
	case wireEventResult:
		return translateResult(raw.Payload), nil, nil
	case wireEventStepFinish:
		return translateStepFinish(raw.Payload), nil, nil
	case wireEventError:
		return []agentbridge.Event{{Kind: agentbridge.EventError, Err: stringField(raw.Payload, "message")}}, nil, nil
	}
	return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "cursor unknown event: " + raw.Type}}, nil, nil
}

func malformedEvent(raw agentbridge.RawEvent) []agentbridge.Event {
	return []agentbridge.Event{{Kind: agentbridge.EventWarning, Text: "malformed cursor stream-json", Err: string(raw.Bytes)}}
}
