package agentbridge

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

const (
	ProgressMessageMetadataCode      = "riido_progress_message_code"
	ProgressMessageMetadataKey       = "riido_progress_message_key"
	ProgressMessageMetadataArgPrefix = "riido_progress_message_arg."
)

type progressMessageTemplate struct {
	Code     int
	Key      string
	Template string
}

type telemetryProgressPayload struct {
	Code    int               `json:"code"`
	Key     string            `json:"key,omitempty"`
	Args    map[string]string `json:"args,omitempty"`
	Message string            `json:"message,omitempty"`
}

var progressPlaceholderPattern = regexp.MustCompile(`\{\{([a-zA-Z0-9_]+)\}\}`)

// Projected from riido-contracts/progressmessage/catalog.ir.riido.json.
var progressMessageTemplates = []progressMessageTemplate{
	{Code: 1001, Key: "agent.thinking", Template: "생각 중. . ."},
	{Code: 1002, Key: "assignment.queued.agent_busy", Template: "지금은 다른 작업을 처리 중이에요. 현재 작업이 끝나는 대로 바로 시작할게요."},
	{Code: 1003, Key: "assignment.stopped.agent_deleted", Template: "에이전트가 삭제되어 진행 중이던 작업이 중지됐어요."},
	{Code: 1004, Key: "assignment.stopped.by_user", Template: "{{actor_name}}님이 직접 종료하였습니다"},
	{Code: 1101, Key: "tool.collecting", Template: "{{label}} 수집 중 - {{description}}"},
	{Code: 1102, Key: "tool.collection_completed_count", Template: "{{label}} 조회 완료 - {{count}}건({{representative_title}} 외)의 요약을 가져왔습니다. . ."},
	{Code: 1103, Key: "tool.running", Template: "{{label}} 실행 중 - {{description}}"},
	{Code: 1104, Key: "tool.completed", Template: "{{label}} 완료 - {{summary}}"},
	{Code: 1201, Key: "assignment.started", Template: "작업을 시작했어요."},
	{Code: 1202, Key: "assignment.completed", Template: "작업을 완료했어요."},
	{Code: 1203, Key: "assignment.failed", Template: "작업을 계속 진행하지 못했어요."},
}

func progressEventFromTelemetryMessage(message string) (Event, bool) {
	message = strings.TrimSpace(message)
	if message == "" {
		return Event{}, false
	}
	if strings.HasPrefix(message, "{") {
		var payload telemetryProgressPayload
		if err := json.Unmarshal([]byte(message), &payload); err == nil && payload.Code > 0 {
			args := cleanProgressArgs(payload.Args)
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
		ProgressMessageMetadataCode: strconv.Itoa(ev.ProgressCode),
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

func renderProgressMessage(code int, args map[string]string) (string, string, bool) {
	for _, item := range progressMessageTemplates {
		if item.Code != code {
			continue
		}
		return renderProgressTemplate(item.Template, args), item.Key, true
	}
	return "", "", false
}

func classifyLegacyProgressMessage(message string) (int, string, map[string]string, bool) {
	switch {
	case message == "생각 중. . ." || message == "생각 중...":
		return 1001, "agent.thinking", nil, true
	case strings.Contains(message, "지금은 다른 작업을 처리 중"):
		return 1002, "assignment.queued.agent_busy", nil, true
	case strings.Contains(message, "에이전트가 삭제되어"):
		return 1003, "assignment.stopped.agent_deleted", nil, true
	}
	if label, description, ok := splitLegacyProgress(message, " 수집 중 - "); ok {
		return 1101, "tool.collecting", map[string]string{"label": label, "description": description}, true
	}
	if label, description, ok := splitLegacyProgress(message, " 실행 중 - "); ok {
		return 1103, "tool.running", map[string]string{"label": label, "description": description}, true
	}
	if label, summary, ok := splitLegacyProgress(message, " 완료 - "); ok {
		return 1104, "tool.completed", map[string]string{"label": label, "summary": summary}, true
	}
	return 0, "", nil, false
}

func splitLegacyProgress(message, marker string) (string, string, bool) {
	idx := strings.Index(message, marker)
	if idx < 0 {
		return "", "", false
	}
	label := strings.TrimSpace(message[:idx])
	detail := strings.TrimSpace(message[idx+len(marker):])
	if label == "" || detail == "" {
		return "", "", false
	}
	return label, detail, true
}

func cleanProgressArgs(args map[string]string) map[string]string {
	if len(args) == 0 {
		return nil
	}
	out := map[string]string{}
	for key, value := range args {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			out[key] = value
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func renderProgressTemplate(template string, args map[string]string) string {
	return progressPlaceholderPattern.ReplaceAllStringFunc(template, func(match string) string {
		parts := progressPlaceholderPattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		value := strings.TrimSpace(args[parts[1]])
		if value == "" {
			value = "not provided"
		}
		return value
	})
}
