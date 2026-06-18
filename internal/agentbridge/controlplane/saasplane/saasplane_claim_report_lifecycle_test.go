package saasplane

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func reportClaimedAssignmentLifecycle(ctx context.Context, plane *Plane, req *bridge.TaskRequest) error {
	if err := plane.StartTask(ctx, req.ID); err != nil {
		return err
	}
	if err := plane.Heartbeat(ctx, controlplane.RuntimeHeartbeat{
		RuntimeID:      RuntimeIDForAgent("daemon-1", AgentBinding{AgentID: "jykim1", RuntimeProvider: "codex"}),
		RunningTaskIDs: []string{req.ID},
	}); err != nil {
		return err
	}
	return reportClaimedAssignmentEvents(ctx, plane, req)
}

func reportClaimedAssignmentEvents(ctx context.Context, plane *Plane, req *bridge.TaskRequest) error {
	events := []agentbridge.Event{
		{Kind: agentbridge.EventProgress, Text: "project go.mod written"},
		{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning},
		{Kind: agentbridge.EventSessionIdentified, SessionID: "sess-1"},
	}
	for _, event := range events {
		if err := plane.ReportEvent(ctx, req.ID, event); err != nil {
			return err
		}
	}
	return plane.CompleteTask(ctx, req.ID, agentbridge.Result{
		Status:    agentbridge.ResultCompleted,
		Output:    "ok",
		SessionID: "sess-1",
	})
}
