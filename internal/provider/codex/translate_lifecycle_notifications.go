package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func codexTurnStartedEvent() []agentbridge.Event {
	return []agentbridge.Event{{
		Kind:  agentbridge.EventLifecycle,
		Phase: agentbridge.StateRunning,
	}}
}

func codexRateLimitsUpdatedEvent() []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventLog,
		Text: "codex rate limits updated",
	}}
}

func codexStructuralLifecycleEvent(method codexMethod) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventLog,
		Text: "codex " + string(method),
	}}
}

func codexUnknownNotificationEvent(method codexMethod) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventLog,
		Text: "codex unknown notification: " + string(method),
	}}
}
