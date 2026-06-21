package supervisor

import (
	"context"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func (a *Actor) appendWorkspaceArchivedEvent(ctx context.Context, task *runningTask, record workdir.ArchiveRecord) {
	if task == nil {
		return
	}
	a.appendTaskWorkspaceEvent(ctx, task, ir.EventWorkdirArchived, eventNativeConfigVersion(task.events), map[string]any{
		"workdirPath": record.WorkdirPath,
		"archiveURI":  record.ArchiveURI,
	})
}
