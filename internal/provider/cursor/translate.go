package cursor

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
)

// Translate maps a Cursor RawEvent to run-scope Events.
//
// Cursor's stream-json shares much of Claude's shape (assistant
// content array, tool_use / tool_result, result event) but has
// quirks: a separate top-level "text" event, "step_finish" carrying
// usage that may need to fall back to the final "result" event, and
// "output_text" instead of "text" inside assistant content.
func Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	switch raw.Source {
	case agentbridge.RawSourceStderr:
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: string(raw.Bytes)}}, nil, nil
	case agentbridge.RawSourceStdout, agentbridge.RawSourceClose:
	}

	switch wireEventType(raw.Type) {
	case wireEventMalformed:
		return []agentbridge.Event{{Kind: agentbridge.EventWarning, Text: "malformed cursor stream-json", Err: string(raw.Bytes)}}, nil, nil

	case wireEventSystem:
		var out []agentbridge.Event
		if sid := stringField(raw.Payload, "session_id"); sid != "" {
			out = append(out, agentbridge.Event{Kind: agentbridge.EventSessionIdentified, SessionID: sid})
		}
		out = append(out, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning})
		return out, nil, nil

	case wireEventText:
		return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: stringField(raw.Payload, "text")}}, nil, nil

	case wireEventAssistant:
		return translateAssistant(raw.Payload), nil, nil

	case wireEventToolUse:
		return []agentbridge.Event{{
			Kind: agentbridge.EventToolCallStarted,
			Tool: agentbridge.ToolRef{
				ID:   stringField(raw.Payload, "id"),
				Name: stringField(raw.Payload, "name"),
				Kind: stringField(raw.Payload, "name"),
				Args: toolargs.FromValue(firstToolInput(raw.Payload)),
			},
		}}, nil, nil

	case wireEventToolResult:
		isErr, _ := raw.Payload["is_error"].(bool)
		kind := agentbridge.EventToolCallCompleted
		if isErr {
			kind = agentbridge.EventToolCallFailed
		}
		return []agentbridge.Event{{
			Kind: kind,
			Tool: agentbridge.ToolRef{ID: stringField(raw.Payload, "tool_use_id")},
		}}, nil, nil

	case wireEventResult:
		return translateResult(raw.Payload), nil, nil

	case wireEventStepFinish:
		if usage, ok := raw.Payload["usage"].(map[string]any); ok {
			return []agentbridge.Event{{Kind: agentbridge.EventUsageDelta, Usage: parseUsage(usage)}}, nil, nil
		}
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "cursor step_finish without usage"}}, nil, nil

	case wireEventError:
		return []agentbridge.Event{{Kind: agentbridge.EventError, Err: stringField(raw.Payload, "message")}}, nil, nil
	}

	return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "cursor unknown event: " + raw.Type}}, nil, nil
}

func translateAssistant(p map[string]any) []agentbridge.Event {
	content, _ := p["content"].([]any)
	var out []agentbridge.Event
	for _, item := range content {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		switch wireContentType(stringField(obj, "type")) {
		case wireContentText, wireContentOutputText:
			out = append(out, agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: stringField(obj, "text")})
		case wireContentThinking:
			out = append(out, agentbridge.Event{Kind: agentbridge.EventThinkingDelta, Text: stringField(obj, "text")})
		case wireContentToolUse:
			out = append(out, agentbridge.Event{
				Kind: agentbridge.EventToolCallStarted,
				Tool: agentbridge.ToolRef{
					ID:   stringField(obj, "id"),
					Name: stringField(obj, "name"),
					Kind: stringField(obj, "name"),
					Args: toolargs.FromValue(firstToolInput(obj)),
				},
			})
		}
	}
	return out
}

func firstToolInput(payload map[string]any) any {
	for _, key := range []string{"input", "tool_input", "args"} {
		if value, ok := payload[key]; ok {
			return value
		}
	}
	return nil
}

func translateResult(p map[string]any) []agentbridge.Event {
	var out []agentbridge.Event
	if usage, ok := p["usage"].(map[string]any); ok {
		out = append(out, agentbridge.Event{Kind: agentbridge.EventUsageDelta, Usage: parseUsage(usage)})
	}
	subtype := wireResultSubtype(stringField(p, "subtype"))
	status := agentbridge.ResultCompleted
	switch subtype {
	case wireResultSubtypeError, wireResultSubtypeExecutionError:
		status = agentbridge.ResultFailed
	case wireResultSubtypeCancelled:
		status = agentbridge.ResultCancelled
	}
	out = append(out, agentbridge.Event{
		Kind: agentbridge.EventResult,
		Result: agentbridge.Result{
			Status: status,
			Output: stringField(p, "result"),
			Error:  stringField(p, "error"),
		},
	})
	return out
}

func parseUsage(m map[string]any) agentbridge.Usage {
	intField := func(k string) int {
		switch v := m[k].(type) {
		case float64:
			return int(v)
		case int:
			return v
		}
		return 0
	}
	return agentbridge.Usage{
		PromptTokens:     intField("input_tokens"),
		CompletionTokens: intField("output_tokens"),
	}
}

func stringField(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	s, _ := m[key].(string)
	return s
}
