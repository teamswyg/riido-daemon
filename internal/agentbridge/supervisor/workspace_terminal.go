package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func (a *Actor) recordTerminalResult(ctx context.Context, running *runningTask, res agentbridge.Result) agentbridge.Result {
	if running == nil {
		return res
	}
	if res.Workdir == "" && running.workspace != nil {
		res.Workdir = running.workspace.Workdir
	}
	a.appendTerminalResultEvent(ctx, running.taskID, running.events, res)
	a.archiveTerminalWorkspace(ctx, running.taskID, running.workspace, running.events, res)
	return res
}

func (a *Actor) archiveTerminalWorkspace(ctx context.Context, taskID string, ws *workdir.Workspace, events *workspaceEventContext, res agentbridge.Result) {
	if ws == nil || a.cfg.Workdir == nil {
		return
	}
	archiver, ok := a.cfg.Workdir.(workdir.Archiver)
	if !ok {
		return
	}
	record, err := archiver.Archive(*ws, workdir.ArchiveRequest{
		ResultStatus: string(res.Status),
		ArchivedAt:   res.FinishedAt,
	})
	if err == nil {
		a.appendWorkspaceArchivedEvent(ctx, taskID, events, record)
		return
	}
	_ = a.cfg.Reporter.ReportEvent(ctx, taskID, agentbridge.Event{
		Kind: agentbridge.EventWarning,
		Text: "workspace archive failed",
		Err:  err.Error(),
	})
}
