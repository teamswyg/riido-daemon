package supervisor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	if res.Status == agentbridge.ResultCompleted && running.execution != nil {
		verifyCtx := ctx
		if running.ctx != nil {
			verifyCtx = running.ctx
		}
		evidence := running.execution.Verify(verifyCtx)
		data, err := json.MarshalIndent(evidence, "", "  ")
		if err == nil && running.workspace != nil {
			err = os.WriteFile(filepath.Join(running.workspace.Artifacts, "verification.json"), data, 0o600)
		}
		res.Workdir = running.execution.Directory
		switch {
		case verifyCtx.Err() != nil:
			res.Status = agentbridge.ResultCancelled
			res.Error = "Execution verification cancelled"
		case evidence.Error != "":
			res.Status = agentbridge.ResultFailed
			res.Error = "Execution verification failed: " + evidence.Error
		case err != nil:
			res.Status = agentbridge.ResultFailed
			res.Error = "Could not persist execution evidence: " + err.Error()
		default:
			res.Output += "\n\nVerified changed files: " + strings.Join(evidence.ChangedFiles, ", ") + fmt.Sprintf("; checks passed: %d", len(evidence.Checks))
		}
	}
	a.appendTaskTerminalResultEvent(ctx, running, res)
	a.archiveTerminalWorkspace(ctx, running, res)
	return res
}

func (a *Actor) archiveTerminalWorkspace(ctx context.Context, task *runningTask, res agentbridge.Result) {
	if task == nil {
		return
	}
	ws := task.workspace
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
		a.appendWorkspaceArchivedEvent(ctx, task, record)
		return
	}
	a.reportTaskEvent(ctx, task, agentbridge.Event{
		Kind: agentbridge.EventWarning,
		Text: "workspace archive failed",
		Err:  err.Error(),
	})
}
