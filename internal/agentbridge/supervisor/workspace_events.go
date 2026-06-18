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

func (a *Actor) appendProviderEvent(ctx context.Context, taskID string, events *workspaceEventContext, ev agentbridge.Event) {
	if events == nil {
		return
	}
	eventType, payload, ok := providerEventDraft(ev)
	if !ok {
		return
	}
	if _, err := events.agentIngestor.Append(ctx, events.draft(eventType, eventNativeConfigVersion(events), payload)); err != nil {
		_ = a.cfg.Reporter.ReportEvent(ctx, taskID, agentbridge.Event{
			Kind: agentbridge.EventWarning,
			Text: "provider event append failed: " + string(eventType),
			Err:  err.Error(),
		})
	}
}

func (a *Actor) appendTerminalResultEvent(ctx context.Context, taskID string, events *workspaceEventContext, res agentbridge.Result) {
	if events == nil {
		return
	}
	eventType, payload := terminalResultDraft(res)
	if _, err := events.ingestor.Append(ctx, events.transitionDraft(eventType, eventNativeConfigVersion(events), payload)); err != nil {
		_ = a.cfg.Reporter.ReportEvent(ctx, taskID, agentbridge.Event{
			Kind: agentbridge.EventWarning,
			Text: "terminal result event append failed: " + string(eventType),
			Err:  err.Error(),
		})
	}
}
