package claude

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
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
		// Bare stderr lines (no JSON) become Log events.
		return []agentbridge.Event{{
			Kind: agentbridge.EventLog,
			Text: string(raw.Bytes),
		}}, nil, nil
	case agentbridge.RawSourceStdout, agentbridge.RawSourceClose:
	}

	switch raw.Type {
	case "malformed":
		return []agentbridge.Event{{
			Kind: agentbridge.EventWarning,
			Text: "malformed claude stream-json line",
			Err:  string(raw.Bytes),
		}}, nil, nil

	case "system":
		return translateSystem(raw), nil, nil

	case "assistant":
		return translateAssistantMessage(raw), nil, nil

	case "user":
		return translateUserMessage(raw), nil, nil

	case "control_request":
		return translateControlRequest(raw), nil, nil

	case "result":
		return translateResult(raw), nil, nil

	case "log":
		return []agentbridge.Event{{
			Kind: agentbridge.EventLog,
			Text: stringField(raw.Payload, "message"),
		}}, nil, nil

	case "error":
		return []agentbridge.Event{{
			Kind: agentbridge.EventError,
			Err:  stringField(raw.Payload, "message"),
		}}, nil, nil

	case "rate_limit_event", "rate_limit":
		// Claude Code emits rate_limit_event when the upstream account is being
		// throttled. It is informational (the CLI keeps the session alive and
		// retries), NOT terminal — surface it as a Warning so the run keeps its
		// semantic-activity timer alive and the UI shows a clear "rate limited"
		// status instead of a generic "unknown event" line.
		return []agentbridge.Event{{
			Kind: agentbridge.EventWarning,
			Text: "claude rate limited",
			Err:  claudeRateLimitDetail(raw.Payload),
		}}, nil, nil

	default:
		// Unknown but well-formed event — surface as Log so the watchdog
		// keeps semantic-activity tracking accurate and we never silently
		// drop something we'll need later (spec §15 item 3).
		return []agentbridge.Event{{
			Kind: agentbridge.EventLog,
			Text: "claude unknown event: " + raw.Type,
		}}, nil, nil
	}
}

func translateSystem(raw agentbridge.RawEvent) []agentbridge.Event {
	var out []agentbridge.Event
	if sid := stringField(raw.Payload, "session_id"); sid != "" {
		out = append(out, agentbridge.Event{Kind: agentbridge.EventSessionIdentified, SessionID: sid})
	}
	out = append(out, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning})
	return out
}

func translateAssistantMessage(raw agentbridge.RawEvent) []agentbridge.Event {
	message, _ := raw.Payload["message"].(map[string]any)
	content, _ := message["content"].([]any)
	var out []agentbridge.Event
	for _, item := range content {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		switch obj["type"] {
		case "text":
			out = append(out, agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: stringField(obj, "text")})
		case "thinking":
			out = append(out, agentbridge.Event{Kind: agentbridge.EventThinkingDelta, Text: stringField(obj, "thinking")})
		case "tool_use":
			out = append(out, agentbridge.Event{
				Kind: agentbridge.EventToolCallStarted,
				Tool: agentbridge.ToolRef{
					ID:   stringField(obj, "id"),
					Name: stringField(obj, "name"),
					Kind: stringField(obj, "name"),
					Args: toolargs.FromValue(obj["input"]),
				},
			})
		}
	}
	return out
}

func translateUserMessage(raw agentbridge.RawEvent) []agentbridge.Event {
	message, _ := raw.Payload["message"].(map[string]any)
	content, _ := message["content"].([]any)
	var out []agentbridge.Event
	for _, item := range content {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if obj["type"] == "tool_result" {
			isErr, _ := obj["is_error"].(bool)
			kind := agentbridge.EventToolCallCompleted
			if isErr {
				kind = agentbridge.EventToolCallFailed
			}
			out = append(out, agentbridge.Event{
				Kind: kind,
				Tool: agentbridge.ToolRef{ID: stringField(obj, "tool_use_id")},
			})
		}
	}
	return out
}

func translateControlRequest(raw agentbridge.RawEvent) []agentbridge.Event {
	request, _ := raw.Payload["request"].(map[string]any)
	if stringField(request, "subtype") == "permission_request" {
		return []agentbridge.Event{{
			Kind: agentbridge.EventToolApprovalNeeded,
			Tool: agentbridge.ToolRef{
				ID:                stringField(request, "tool_use_id"),
				Name:              stringField(request, "tool_name"),
				Kind:              stringField(request, "tool_name"),
				Args:              toolargs.FromValue(firstToolInput(request)),
				ProviderRequestID: stringField(raw.Payload, "request_id"),
			},
		}}
	}
	// Don't silently drop unknown control requests (spec §15 item 3).
	return []agentbridge.Event{{
		Kind: agentbridge.EventLog,
		Text: "claude unknown control_request subtype: " + stringField(request, "subtype"),
	}}
}
