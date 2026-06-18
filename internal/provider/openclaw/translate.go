package openclaw

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Translate maps an OpenClaw RawEvent into run-scope Events.
func Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Source == agentbridge.RawSourceStderr {
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: string(raw.Bytes)}}, nil, nil
	}
	switch {
	case wireFrameType(raw.Type) == wireFrameMalformed:
		return []agentbridge.Event{{Kind: agentbridge.EventWarning, Text: "malformed openclaw output", Err: string(raw.Bytes)}}, nil, nil
	case wireFrameType(raw.Type) == wireFrameFullResult:
		return translateFullResult(raw.Payload), nil, nil
	case strings.HasPrefix(raw.Type, wireFrameNDJSONPrefix):
		event := wireNDJSONEvent(strings.TrimPrefix(raw.Type, wireFrameNDJSONPrefix))
		return translateNDJSON(event, raw.Payload), nil, nil
	}
	return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "openclaw unknown frame: " + raw.Type}}, nil, nil
}
