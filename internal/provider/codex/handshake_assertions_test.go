package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func hasSessionIdentified(events []agentbridge.Event, sessionID string) bool {
	for _, ev := range events {
		if ev.Kind == agentbridge.EventSessionIdentified && ev.SessionID == sessionID {
			return true
		}
	}
	return false
}

func hasLifecycleRunning(events []agentbridge.Event) bool {
	for _, ev := range events {
		if ev.Kind == agentbridge.EventLifecycle && ev.Phase == agentbridge.StateRunning {
			return true
		}
	}
	return false
}

func hasTextDelta(events []agentbridge.Event, text string) bool {
	for _, ev := range events {
		if ev.Kind == agentbridge.EventTextDelta && ev.Text == text {
			return true
		}
	}
	return false
}

func hasToolApproval(events []agentbridge.Event, commandID string) bool {
	for _, ev := range events {
		if ev.Kind == agentbridge.EventToolApprovalNeeded && ev.Tool.ID == commandID {
			return true
		}
	}
	return false
}

func hasCompletedResult(events []agentbridge.Event) bool {
	for _, ev := range events {
		if ev.Kind == agentbridge.EventResult && ev.Result.Status == agentbridge.ResultCompleted {
			return true
		}
	}
	return false
}
