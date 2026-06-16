package agentbridge

import (
	"encoding/json"
	"strconv"
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

func ProgressEventMetadata(ev Event) map[string]string {
	if ev.ProgressCode <= 0 {
		return nil
	}
	metadata := map[string]string{
		ProgressMessageMetadataCode: strconv.Itoa(int(ev.ProgressCode)),
	}
	if strings.TrimSpace(ev.ProgressKey) != "" {
		metadata[ProgressMessageMetadataKey] = strings.TrimSpace(ev.ProgressKey)
	}
	for key, value := range ev.ProgressArgs {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		metadata[ProgressMessageMetadataArgPrefix+key] = value
	}
	return metadata
}

func renderProgressMessage(code ProgressCode, args map[string]string) (string, string, bool) {
	args = progressmessage.NormalizeArgsForCode(int(code), args)
	rendered, ok := progressmessage.Render(int(code), args, progressmessage.DefaultLocale)
	if !ok {
		return "", "", false
	}
	return rendered, progressMessageKey(code), true
}

func cleanProgressArgs(args map[string]any) map[string]string {
	if len(args) == 0 {
		return nil
	}
	out := map[string]string{}
	for key, value := range args {
		key = strings.TrimSpace(key)
		rendered := strings.TrimSpace(progressArgString(value))
		if key != "" && rendered != "" {
			out[key] = rendered
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
