package main

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func eventKindConstName(kind agentbridge.EventKind) string {
	switch kind {
	case agentbridge.EventCancellation:
		return "EventCancellation"
	case agentbridge.EventProcessExit:
		return "EventProcessExit"
	default:
		return "Event" + camel(string(kind))
	}
}
