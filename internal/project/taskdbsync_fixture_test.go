package project

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func sampleTaskDBState(t *testing.T) StateFile {
	t.Helper()

	projection, err := FromMwsdSnapshot(sampleSnapshot())
	if err != nil {
		t.Fatalf("FromMwsdSnapshot returned error: %v", err)
	}

	return StateFromProjection(projection)
}

func taskDBSyncTime(minute int) time.Time {
	return time.Date(2026, 5, 20, 8, minute, 0, 0, time.UTC)
}

func findTaskRecord(t *testing.T, db taskdb.TaskDB, taskID string) taskdb.TaskRecord {
	t.Helper()

	for _, record := range db.Tasks {
		if record.ID == taskID {
			return record
		}
	}

	t.Fatalf("task %s not found", taskID)
	return taskdb.TaskRecord{}
}
