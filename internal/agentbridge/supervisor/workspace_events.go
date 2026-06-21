package supervisor

import (
	"context"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) appendWorkspaceEvent(ctx context.Context, taskID string, events *workspaceEventContext, eventType ir.EventType, nativeConfigVersion string, payload map[string]any) {
	if events == nil {
		return
	}
	if _, err := events.ingestor.Append(ctx, events.draft(eventType, nativeConfigVersion, payload)); err != nil {
		_ = a.cfg.Reporter.ReportEvent(ctx, taskID, agentbridge.Event{
			Kind: agentbridge.EventWarning,
			Text: "workspace event append failed: " + string(eventType),
			Err:  err.Error(),
		})
	}
}
