package supervisor

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestProviderEventDraftMapsCatCEvents(t *testing.T) {
	for _, tc := range providerEventDraftMappingCases() {
		t.Run(tc.name, func(t *testing.T) {
			got, payload, ok := providerEventDraft(tc.ev)
			if !ok {
				t.Fatalf("expected mapping for %+v", tc.ev)
			}
			if got != tc.want {
				t.Fatalf("event type = %s, want %s", got, tc.want)
			}
			if len(payload) == 0 {
				t.Fatalf("payload must not be empty")
			}
		})
	}
}

func TestProviderEventDraftSkipsTerminalResult(t *testing.T) {
	if _, _, ok := providerEventDraft(agentbridge.Event{Kind: agentbridge.EventResult}); ok {
		t.Fatal("EventResult must stay outside non-transition Cat C mapping")
	}
}

func providerEventDraftMappingCases() []struct {
	name string
	ev   agentbridge.Event
	want ir.EventType
} {
	return []struct {
		name string
		ev   agentbridge.Event
		want ir.EventType
	}{
		{"session", agentbridge.Event{Kind: agentbridge.EventSessionIdentified, SessionID: "s-1"}, ir.EventSessionPinned},
		{"text", agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "hello"}, ir.EventTextDelta},
		{"thinking", agentbridge.Event{Kind: agentbridge.EventThinkingDelta, Text: "why"}, ir.EventReasoningDelta},
		{"tool-start", providerToolEvent(agentbridge.EventToolCallStarted), ir.EventToolCallStarted},
		{"tool-done", providerToolEvent(agentbridge.EventToolCallCompleted), ir.EventToolCallFinished},
		{"approval", providerApprovalEvent(), ir.EventApprovalRequested},
		{"usage", agentbridge.Event{Kind: agentbridge.EventUsageDelta, Usage: agentbridge.Usage{PromptTokens: 1}}, ir.EventUsageDelta},
		{"warning", agentbridge.Event{Kind: agentbridge.EventWarning, Text: "careful"}, ir.EventLogLine},
	}
}

func providerToolEvent(kind agentbridge.EventKind) agentbridge.Event {
	return agentbridge.Event{
		Kind: kind,
		Tool: agentbridge.ToolRef{ID: "tool-1", Name: "bash"},
	}
}

func providerApprovalEvent() agentbridge.Event {
	return agentbridge.Event{
		Kind: agentbridge.EventToolApprovalNeeded,
		Tool: agentbridge.ToolRef{ID: "approval-1", Kind: "exec"},
	}
}
