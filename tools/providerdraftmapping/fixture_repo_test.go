package main

func sourceFixture() string {
	return `package supervisor
func providerEventDraft(ev agentbridge.Event) (ir.EventType, map[string]any, bool) {
	switch ev.Kind {
	case agentbridge.EventTextDelta:
		return ir.EventTextDelta, map[string]any{"text": ev.Text}, true
	default:
		return "", nil, false
	}
}`
}

func fixtureSkippedEvents(except string) []SkippedEvent {
	var out []SkippedEvent
	for kind := range runtimeEventKinds() {
		if kind != except {
			out = append(out, SkippedEvent{EventKind: kind, EventKindConst: eventKindConstForValue(kind), Reason: "fixture skip"})
		}
	}
	return out
}

func eventKindConstForValue(kind string) string {
	for name, value := range eventKindByConst() {
		if value == kind {
			return name
		}
	}
	return ""
}
