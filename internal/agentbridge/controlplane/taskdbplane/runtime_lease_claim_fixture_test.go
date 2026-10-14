package taskdbplane

import (
	"strconv"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func leaseClaimNow() time.Time {
	return time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
}

func runtimeLeaseFixture(taskID, runtimeID, fingerprint string, token int64, now time.Time) RuntimeLeaseRecord {
	return RuntimeLeaseRecord{
		LeaseID:               "runtime-lease:" + taskID + ":" + strconv.FormatInt(token, 10),
		TaskID:                taskID,
		RuntimeID:             runtimeID,
		CapabilityFingerprint: fingerprint,
		ClaimedAt:             now.Add(-2 * time.Hour),
		LeaseUntil:            now.Add(-time.Hour),
		FencingToken:          token,
	}
}

func runningCodexTaskDB() taskdb.TaskDB {
	db := queuedCodexTaskDB()
	db.Tasks[0].State = task.StateRunning
	return db
}

func assertSingleLease(tb testing.TB, registry RuntimeLeaseRegistry) RuntimeLeaseRecord {
	tb.Helper()
	if len(registry.Leases) != 1 {
		tb.Fatalf("lease count = %d, want 1: %+v", len(registry.Leases), registry.Leases)
	}
	return registry.Leases[0]
}
