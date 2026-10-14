package agentbridge

import (
	"encoding/json"
	"strings"

	"github.com/teamswyg/riido-contracts/progressmessage"
)

func progressEventFromTelemetryMessage(message string) (Event, bool) {
	message = strings.TrimSpace(message)
	if message == "" {
		return Event{}, false
	}
	if strings.HasPrefix(message, "{") {
		var payload telemetryProgressPayload
		if err := json.Unmarshal([]byte(message), &payload); err == nil && payload.Code > 0 {
			args := cleanProgressArgs(payload.Args)
			args = progressmessage.NormalizeArgsForCode(int(payload.Code), args)
			text := strings.TrimSpace(payload.Message)
			key := strings.TrimSpace(payload.Key)
			if rendered, renderedKey, ok := renderProgressMessage(payload.Code, args); ok {
				text = rendered
				if key == "" {
					key = renderedKey
				}
			}
			if text == "" {
				text = message
			}
			return Event{Kind: EventProgress, Text: text, ProgressCode: payload.Code, ProgressKey: key, ProgressArgs: args}, true
		}
	}
	if code, key, args, ok := classifyLegacyProgressMessage(message); ok {
		args = progressmessage.NormalizeArgsForCode(int(code), args)
		text := message
		if rendered, renderedKey, ok := renderProgressMessage(code, args); ok {
			text = rendered
			if key == "" {
				key = renderedKey
			}
		}
		return Event{Kind: EventProgress, Text: text, ProgressCode: code, ProgressKey: key, ProgressArgs: args}, true
	}
	return Event{Kind: EventProgress, Text: message}, true
}
