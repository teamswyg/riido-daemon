package claude

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Translate maps a Claude RawEvent to zero or more provider-neutral
// run-scope Events (and optionally a Command, though Claude's stream-json
// generally doesn't require imperative output beyond the reducer's own
// Approve/Cancel responses).
//
// Reference: docs/20-domain/provider-runtime.md and Anthropic stream-json docs.
//
// Translate is a pure function; state is carried by the reducer.
func Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	switch raw.Source {
	case agentbridge.RawSourceStderr:
		return translateStderrRaw(raw), nil, nil
	case agentbridge.RawSourceStdout, agentbridge.RawSourceClose:
	}

	switch wireEventType(raw.Type) {
	case wireEventMalformed:
		return translateMalformed(raw), nil, nil

	case wireEventSystem:
		return translateSystem(raw), nil, nil

	case wireEventAssistant:
		return translateAssistantMessage(raw), nil, nil

	case wireEventUser:
		return translateUserMessage(raw), nil, nil

	case wireEventControl:
		return translateControlRequest(raw), nil, nil

	case wireEventResult:
		return translateResult(raw), nil, nil

	case wireEventLog:
		return translateLog(raw), nil, nil

	case wireEventError:
		return translateError(raw), nil, nil

	case wireEventRateLimitAlt, wireEventRateLimit:
		return translateRateLimit(raw), nil, nil

	default:
		return translateUnknown(raw), nil, nil
	}
}
