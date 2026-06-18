package supervisor

import (
	"context"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func (a *Actor) appendWorkspaceArchivedEvent(ctx context.Context, taskID string, events *workspaceEventContext, record workdir.ArchiveRecord) {
	a.appendWorkspaceEvent(ctx, taskID, events, ir.EventWorkdirArchived, eventNativeConfigVersion(events), map[string]any{
		"workdirPath": record.WorkdirPath,
		"archiveURI":  record.ArchiveURI,
	})
}
