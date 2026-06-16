package agentbridge

import (
	"encoding/json"
	"maps"
	"strconv"
	"strings"
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
			args = normalizeProgressArgsForCode(payload.Code, args)
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
		args = normalizeProgressArgsForCode(code, args)
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
	for _, item := range progressMessageTemplates {
		if item.Code != code {
			continue
		}
		return renderProgressTemplate(item.Template, normalizeProgressArgsForCode(code, args)), item.Key, true
	}
	return "", "", false
}

func normalizeProgressArgsForCode(code ProgressCode, args map[string]string) map[string]string {
	if len(args) == 0 {
		return args
	}
	label := strings.TrimSpace(args["label"])
	if label == "" {
		return args
	}
	normalized := normalizeProgressLabelForCode(code, label)
	if normalized == label {
		return args
	}
	out := make(map[string]string, len(args))
	maps.Copy(out, args)
	out["label"] = normalized
	return out
}

func normalizeProgressLabelForCode(code ProgressCode, label string) string {
	switch code {
	case ProgressCodeToolCollecting:
		return trimProgressLabelSuffixes(label, " 수집 중", " 수집", " 조회 중", " 조회")
	case ProgressCodeToolCollectionCompletedCount:
		return trimProgressLabelSuffixes(label, " 조회 완료", " 완료", " 조회")
	case ProgressCodeToolRunning:
		return trimProgressLabelSuffixes(label, " 실행 중", " 진행 중", " 처리 중", " 실행", " 진행", " 처리")
	case ProgressCodeToolCompleted:
		return trimProgressLabelSuffixes(label, " 조회 완료", " 실행 완료", " 진행 완료", " 처리 완료", " 완료됨", " 완료", " 종료", " 끝남")
	default:
		return strings.TrimSpace(label)
	}
}

func trimProgressLabelSuffixes(label string, suffixes ...string) string {
	label = strings.TrimSpace(label)
	for {
		changed := false
		for _, suffix := range suffixes {
			if !strings.HasSuffix(label, suffix) {
				continue
			}
			next := strings.TrimSpace(strings.TrimSuffix(label, suffix))
			if next == "" {
				continue
			}
			label = next
			changed = true
			break
		}
		if !changed {
			return label
		}
	}
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
