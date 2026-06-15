package saasplane

import (
	"context"
	"strings"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneClaimsAndReportsAssignment(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:               "asn-1",
		TaskID:           "task-a",
		ComponentID:      "component-1",
		AgentID:          "jykim1",
		RuntimeProvider:  "codex",
		Prompt:           "golang hello world quickly",
		AgentInstruction: "write concise Korean progress updates",
		State:            assignmentcontract.AssignmentQueued,
		LeaseToken:       "lease-1",
	})
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req == nil || req.ID != "asn-1" || req.Provider != "codex" {
		t.Fatalf("request = %+v", req)
	}
	if got := req.Metadata[MetadataAssignmentID]; got != "asn-1" {
		t.Fatalf("assignment metadata = %q", got)
	}
	if got := req.Metadata[controlplane.MetadataTaskID]; got != "task-a" {
		t.Fatalf("task metadata = %q", got)
	}
	if got := req.Metadata["workspace_id"]; got != "component-1" {
		t.Fatalf("workspace_id = %q", got)
	}
	if !strings.Contains(req.Prompt, "<riido_log>") || !strings.Contains(req.Prompt, "golang hello world") || !strings.Contains(req.Prompt, "write concise Korean progress updates") {
		t.Fatalf("prompt missing telemetry contract: %q", req.Prompt)
	}
	if got := req.Metadata[agentbridge.MetadataTelemetryContract]; got != agentbridge.TelemetryPlacementPrompt {
		t.Fatalf("telemetry placement = %q", got)
	}
	if got := req.Metadata[agentbridge.MetadataAgentInstruction]; got != agentbridge.TelemetryPlacementPrompt {
		t.Fatalf("instruction placement = %q", got)
	}

	if err := plane.StartTask(context.Background(), req.ID); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID:      RuntimeIDForAgent("daemon-1", AgentBinding{AgentID: "jykim1", RuntimeProvider: "codex"}),
		RunningTaskIDs: []string{req.ID},
	}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	if err := plane.ReportEvent(context.Background(), req.ID, agentbridge.Event{Kind: agentbridge.EventProgress, Text: "project go.mod written"}); err != nil {
		t.Fatalf("ReportEvent: %v", err)
	}
	if err := plane.ReportEvent(context.Background(), req.ID, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}); err != nil {
		t.Fatalf("ReportEvent running: %v", err)
	}
	if err := plane.CompleteTask(context.Background(), req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "ok"}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	fake.assertEvent(t, assignmentcontract.EventAssignmentReady)
	fake.assertEvent(t, assignmentcontract.EventRiidoLog)
	fake.assertEvent(t, assignmentcontract.EventAssignmentRunning)
	fake.assertEvent(t, assignmentcontract.EventAssignmentCompleted)
	heartbeats := fake.heartbeatsFor("jykim1")
	if len(heartbeats) != 1 || len(heartbeats[0].ActiveAssignmentIDs) != 1 || heartbeats[0].ActiveAssignmentIDs[0] != "asn-1" {
		t.Fatalf("heartbeats = %+v", heartbeats)
	}
}

func TestTaskRequestPlacesTelemetryForSystemPromptProviders(t *testing.T) {
	assignment := assignmentcontract.Assignment{
		ID:                       "asn-1",
		TaskID:                   "task-a",
		ComponentID:              "component-1",
		AgentID:                  "jykim1",
		RuntimeProvider:          "claude",
		ModelID:                  "claude-opus-4-7",
		Prompt:                   "golang hello world quickly",
		AgentInstruction:         "act as a backend reviewer",
		AllowExperimentalRuntime: true,
	}
	req := taskRequestFromAssignment(assignment)
	if req.Prompt != assignment.Prompt {
		t.Fatalf("claude prompt should remain user task only: %q", req.Prompt)
	}
	if !strings.Contains(req.SystemPrompt, "<riido_log>") || !strings.Contains(req.SystemPrompt, "act as a backend reviewer") {
		t.Fatalf("claude system prompt missing runtime instructions: %q", req.SystemPrompt)
	}
	if got := req.Metadata[agentbridge.MetadataTelemetryContract]; got != agentbridge.TelemetryPlacementSystemPrompt {
		t.Fatalf("telemetry placement = %q", got)
	}
	if got := req.Metadata[agentbridge.MetadataAgentInstruction]; got != agentbridge.TelemetryPlacementSystemPrompt {
		t.Fatalf("instruction placement = %q", got)
	}
	if !req.AllowExperimentalRuntime {
		t.Fatalf("allow experimental runtime was not copied from assignment")
	}
	if req.Model != assignment.ModelID {
		t.Fatalf("model_id was not copied from assignment: %q", req.Model)
	}
	if got := req.Metadata[MetadataModelID]; got != assignment.ModelID {
		t.Fatalf("metadata model_id = %q, want %q", got, assignment.ModelID)
	}
}

func TestPlaneReportsStructuredProgressMetadata(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "ship it",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	})
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if err := plane.ReportEvent(context.Background(), req.ID, agentbridge.Event{
		Kind:         agentbridge.EventProgress,
		Text:         "생각 중. . .",
		ProgressCode: 1001,
		ProgressKey:  "agent.thinking",
	}); err != nil {
		t.Fatalf("ReportEvent: %v", err)
	}
	if len(fake.events) != 1 {
		t.Fatalf("events = %+v", fake.events)
	}
	event := fake.events[0]
	if event.Metadata[agentbridge.ProgressMessageMetadataCode] != "1001" ||
		event.Metadata[agentbridge.ProgressMessageMetadataKey] != "agent.thinking" {
		t.Fatalf("metadata = %+v", event.Metadata)
	}
}
