package supervisor

import (
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func providerEventDraft(ev agentbridge.Event) (ir.EventType, map[string]any, bool) {
	switch ev.Kind {
	case agentbridge.EventLifecycle:
		return ir.EventStatusUpdate, map[string]any{
			"text":  "provider lifecycle update",
			"phase": string(ev.Phase),
		}, true
	case agentbridge.EventSessionIdentified:
		return ir.EventSessionPinned, map[string]any{"providerSessionID": ev.SessionID}, true
	case agentbridge.EventTextDelta:
		return ir.EventTextDelta, map[string]any{"text": ev.Text}, true
	case agentbridge.EventThinkingDelta:
		return ir.EventReasoningDelta, map[string]any{"text": ev.Text, "private": true}, true
	case agentbridge.EventToolCallStarted:
		return ir.EventToolCallStarted, toolPayload(ev.Tool), true
	case agentbridge.EventToolCallCompleted:
		payload := toolPayload(ev.Tool)
		payload["result"] = "completed"
		return ir.EventToolCallFinished, payload, true
	case agentbridge.EventToolCallFailed:
		payload := toolPayload(ev.Tool)
		payload["error"] = ev.Err
		return ir.EventToolCallFinished, payload, true
	case agentbridge.EventToolApprovalNeeded:
		return ir.EventApprovalRequested, map[string]any{
			"approvalID": ev.Tool.ID,
			"kind":       textutil.FirstNonEmpty(ev.Tool.Kind, "tool"),
			"payload":    toolPayload(ev.Tool),
		}, true
	case agentbridge.EventUsageDelta:
		return ir.EventUsageDelta, map[string]any{"usage": usagePayload(ev.Usage)}, true
	case agentbridge.EventLog:
		return ir.EventLogLine, logLinePayload("info", ev), true
	case agentbridge.EventWarning:
		return ir.EventLogLine, logLinePayload("warning", ev), true
	case agentbridge.EventError:
		return ir.EventLogLine, logLinePayload("error", ev), true
	default:
		return "", nil, false
	}
}

func logLinePayload(level string, ev agentbridge.Event) map[string]any {
	payload := map[string]any{"level": level, "text": ev.Text}
	if ev.Err != "" {
		payload["error"] = ev.Err
	}
	return payload
}
