package supervisor

import (
	"context"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/ir/ingest"
	"github.com/teamswyg/riido-daemon/internal/workdir"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func (a *Actor) newWorkspaceEventContext(ws workdir.Workspace, statusRuntimeID string, req *bridge.TaskRequest, logicalTaskID, runID string, capView runtimeactor.Capability) (*workspaceEventContext, error) {
	sink, err := workdir.NewRunEventSink(ws)
	if err != nil {
		return nil, err
	}
	ingestor, err := ingest.New(ingest.Config{
		Sink:                sink,
		RiidoDaemonVersion:  a.cfg.RiidoDaemonVersion,
		PolicyBundleVersion: a.cfg.PolicyBundleVersion,
		ActorKind:           ir.ActorDaemon,
		ActorID:             a.cfg.DaemonID,
	})
	if err != nil {
		return nil, err
	}
	agentIngestor, err := ingest.New(ingest.Config{
		Sink:                sink,
		RiidoDaemonVersion:  a.cfg.RiidoDaemonVersion,
		PolicyBundleVersion: a.cfg.PolicyBundleVersion,
		ActorKind:           ir.ActorAgent,
		ActorID:             runID,
	})
	if err != nil {
		return nil, err
	}
	return &workspaceEventContext{
		taskID:        logicalTaskID,
		runID:         runID,
		runtimeID:     statusRuntimeID,
		capability:    capView,
		ingestor:      ingestor,
		agentIngestor: agentIngestor,
	}, nil
}

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

func providerEventDraft(ev agentbridge.Event) (ir.EventType, map[string]any, bool) {
	switch ev.Kind {
	case agentbridge.EventLifecycle:
		return ir.EventStatusUpdate, map[string]any{
			"text":  "provider lifecycle update",
			"phase": string(ev.Phase),
		}, true
	case agentbridge.EventSessionIdentified:
		return ir.EventSessionPinned, map[string]any{
			"providerSessionID": ev.SessionID,
		}, true
	case agentbridge.EventTextDelta:
		return ir.EventTextDelta, map[string]any{
			"text": ev.Text,
		}, true
	case agentbridge.EventThinkingDelta:
		return ir.EventReasoningDelta, map[string]any{
			"text":    ev.Text,
			"private": true,
		}, true
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
		return ir.EventUsageDelta, map[string]any{
			"usage": usagePayload(ev.Usage),
		}, true
	case agentbridge.EventLog:
		return ir.EventLogLine, map[string]any{
			"level": "info",
			"text":  ev.Text,
		}, true
	case agentbridge.EventWarning:
		return ir.EventLogLine, map[string]any{
			"level": "warning",
			"text":  ev.Text,
			"error": ev.Err,
		}, true
	case agentbridge.EventError:
		return ir.EventLogLine, map[string]any{
			"level": "error",
			"text":  ev.Text,
			"error": ev.Err,
		}, true
	default:
		return "", nil, false
	}
}
