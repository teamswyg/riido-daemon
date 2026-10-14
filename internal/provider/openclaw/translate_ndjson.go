package openclaw

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func translateNDJSON(event wireNDJSONEvent, p map[string]any) []agentbridge.Event {
	switch event {
	case wireNDJSONText:
		return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: stringField(p, "text")}}
	case wireNDJSONLog:
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: stringField(p, "message")}}
	case wireNDJSONError:
		return []agentbridge.Event{{Kind: agentbridge.EventError, Err: stringField(p, "message")}}
	case wireNDJSONSession:
		return []agentbridge.Event{{Kind: agentbridge.EventSessionIdentified, SessionID: stringField(p, "session_id")}}
	case wireNDJSONUsage:
		return []agentbridge.Event{{Kind: agentbridge.EventUsageDelta, Usage: parseUsage(p)}}
	default:
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "openclaw ndjson unknown event: " + string(event)}}
	}
}
