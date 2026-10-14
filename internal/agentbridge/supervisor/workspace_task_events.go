package supervisor

import (
	"context"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) appendTaskWorkspaceEvent(
	ctx context.Context,
	task *runningTask,
	eventType ir.EventType,
	nativeConfigVersion string,
	payload map[string]any,
) {
	if task == nil || task.events == nil {
		return
	}
	if _, err := task.events.ingestor.Append(ctx, task.events.draft(eventType, nativeConfigVersion, payload)); err != nil {
		a.reportTaskEvent(ctx, task, workspaceAppendWarning("workspace event", eventType, err))
	}
}

func (a *Actor) appendTaskProviderEvent(ctx context.Context, task *runningTask, ev agentbridge.Event) {
	if task == nil || task.events == nil {
		return
	}
	eventType, payload, ok := providerEventDraft(ev)
	if !ok {
		return
	}
	nativeConfigVersion := eventNativeConfigVersion(task.events)
	if _, err := task.events.agentIngestor.Append(ctx, task.events.draft(eventType, nativeConfigVersion, payload)); err != nil {
		a.reportTaskEvent(ctx, task, workspaceAppendWarning("provider event", eventType, err))
	}
}

func (a *Actor) appendTaskTerminalResultEvent(ctx context.Context, task *runningTask, res agentbridge.Result) {
	if task == nil || task.events == nil {
		return
	}
	eventType, payload := terminalResultDraft(res)
	nativeConfigVersion := eventNativeConfigVersion(task.events)
	draft := task.events.transitionDraft(eventType, nativeConfigVersion, payload)
	if _, err := task.events.ingestor.Append(ctx, draft); err != nil {
		a.reportTaskEvent(ctx, task, workspaceAppendWarning("terminal result event", eventType, err))
	}
}

func workspaceAppendWarning(label string, eventType ir.EventType, err error) agentbridge.Event {
	return agentbridge.Event{
		Kind: agentbridge.EventWarning,
		Text: label + " append failed: " + string(eventType),
		Err:  err.Error(),
	}
}
