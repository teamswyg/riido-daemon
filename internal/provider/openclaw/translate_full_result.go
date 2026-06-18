package openclaw

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

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
	out = append(out, agentbridge.Event{
		Kind:   agentbridge.EventResult,
		Result: fullResultOutcome(p, text),
	})
	return out
}

func fullResultOutcome(p map[string]any, text string) agentbridge.Result {
	errMsg := stringField(p, "error")
	status := agentbridge.ResultCompleted
	if errMsg != "" {
		status = agentbridge.ResultFailed
	} else if text == "" {
		status = agentbridge.ResultFailed
		errMsg = "openclaw full_result completed without text payload"
	}
	return agentbridge.Result{Status: status, Output: text, Error: errMsg}
}
