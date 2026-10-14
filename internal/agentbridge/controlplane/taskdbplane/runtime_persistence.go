package taskdbplane

import (
	"context"
	"strings"

	c9lock "github.com/teamswyg/riido-daemon/internal/lock"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
)

func taskDBChanged(before, after taskdb.TaskDB, taskID string) bool {
	if len(before.Transitions) != len(after.Transitions) || len(before.CommandReceipts) != len(after.CommandReceipts) {
		return true
	}
	beforeRecord, beforeOK := findTask(before, taskID)
	afterRecord, afterOK := findTask(after, taskID)
	if beforeOK != afterOK {
		return true
	}
	if !beforeOK {
		return false
	}
	return beforeRecord.State != afterRecord.State ||
		beforeRecord.UpdatedAt != afterRecord.UpdatedAt ||
		beforeRecord.TransitionCount != afterRecord.TransitionCount ||
		beforeRecord.CommandReceiptCount != afterRecord.CommandReceiptCount
}

func runtimeRegistryPath(taskDBPath string) string {
	if before, ok := strings.CutSuffix(taskDBPath, ".json"); ok {
		return before + ".runtimes.json"
	}
	return taskDBPath + ".runtimes.json"
}

func runtimeLeaseRegistryPath(taskDBPath string) string {
	if before, ok := strings.CutSuffix(taskDBPath, ".json"); ok {
		return before + ".leases.json"
	}
	return taskDBPath + ".leases.json"
}

func (p *Plane) withFileLock(ctx context.Context, fn func() error) error {
	return c9lock.WithFile(ctx, p.lockPath, fn)
}

func writeJSONAtomic(path string, value any) error {
	if err := fileutil.WriteJSONAtomic(path, value); err != nil {
		return planeWrapf(ErrTaskDBPlanePersistence, "json.write", err, "write JSON %s", path)
	}
	return nil
}
