package openclaw

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Translate maps an OpenClaw RawEvent into run-scope Events.
//
// Two RawEvent shapes (set by Parser):
//   - Type "full_result"  → one big JSON object: session_id, text,
//     usage, optional error. We emit SessionIdentified + UsageDelta +
//     TextDelta (if any) + Result(completed|failed).
//   - Type "ndjson:<event>" → streaming event. Recognized events:
//     text, log, error, session.
//
// Anything else → Log so we don't silently drop unfamiliar payloads.
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

func translateFullResult(p map[string]any) []agentbridge.Event {
	var out []agentbridge.Event
	if sid := fullResultSessionID(p); sid != "" {
		out = append(out, agentbridge.Event{Kind: agentbridge.EventSessionIdentified, SessionID: sid})
	}
	if usage, ok := fullResultUsage(p); ok {
		out = append(out, agentbridge.Event{Kind: agentbridge.EventUsageDelta, Usage: parseUsage(usage)})
	}
	text := fullResultText(p)
	if text != "" {
		out = append(out, agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: text})
	}
	errMsg := stringField(p, "error")
	status := agentbridge.ResultCompleted
	if errMsg != "" {
		status = agentbridge.ResultFailed
	} else if text == "" {
		status = agentbridge.ResultFailed
		errMsg = "openclaw full_result completed without text payload"
	}
	out = append(out, agentbridge.Event{
		Kind: agentbridge.EventResult,
		Result: agentbridge.Result{
			Status: status,
			Output: text,
			Error:  errMsg,
		},
	})
	return out
}

func fullResultSessionID(p map[string]any) string {
	if sid := stringField(p, "session_id"); sid != "" {
		return sid
	}
	if meta, ok := mapField(p, "meta"); ok {
		if agentMeta, ok := mapField(meta, "agentMeta"); ok {
			return stringField(agentMeta, "sessionId")
		}
	}
	return ""
}

func fullResultUsage(p map[string]any) (map[string]any, bool) {
	if usage, ok := mapField(p, "usage"); ok {
		return usage, true
	}
	if meta, ok := mapField(p, "meta"); ok {
		if agentMeta, ok := mapField(meta, "agentMeta"); ok {
			if usage, ok := mapField(agentMeta, "usage"); ok {
				return usage, true
			}
			if usage, ok := mapField(agentMeta, "lastCallUsage"); ok {
				return usage, true
			}
		}
	}
	return nil, false
}

func fullResultText(p map[string]any) string {
	if text := stringField(p, "text"); text != "" {
		return text
	}
	payloads, _ := p["payloads"].([]any)
	for _, payload := range payloads {
		m, ok := payload.(map[string]any)
		if !ok {
			continue
		}
		if text := stringField(m, "text"); text != "" {
			return text
		}
	}
	return ""
}

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
	}
	return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "openclaw ndjson unknown event: " + string(event)}}
}

func parseUsage(m map[string]any) agentbridge.Usage {
	intField := func(keys ...string) int {
		for _, k := range keys {
			if v := intValue(m[k]); v != 0 {
				return v
			}
		}
		return 0
	}
	return agentbridge.Usage{
		PromptTokens:     intField("prompt_tokens", "input"),
		CompletionTokens: intField("completion_tokens", "output"),
		ReasoningTokens:  intField("reasoning_tokens", "reasoning"),
		CacheReadTokens:  intField("cache_read_tokens", "cacheRead"),
		CacheWriteTokens: intField("cache_write_tokens", "cacheWrite"),
	}
}

func intValue(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

func mapField(m map[string]any, key string) (map[string]any, bool) {
	if m == nil {
		return nil, false
	}
	child, ok := m[key].(map[string]any)
	return child, ok
}

func stringField(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	s, _ := m[key].(string)
	return s
}
