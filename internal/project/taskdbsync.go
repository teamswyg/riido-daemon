package project

import (
	"sort"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

// SyncTaskDBFromState projects the workspace task source into the local task DB
// without taking ownership of guarded task mutation. Runtime mutations remain
// owned by internal/taskdb.
func SyncTaskDBFromState(existing taskdb.TaskDB, state StateFile, now time.Time) taskdb.TaskDB {
	db := normalizeTaskDB(existing)
	stamp := timestamp(now)
	db.ProjectionVersion = state.ProjectionVersion
	db.Root = state.Root
	db.Domain = state.Domain
	db.UpdatedAt = stamp
	db.RecommendedProvider = state.RecommendedProvider
	db.RecommendedDecisionLLM = state.RecommendedDecisionLLM
	db.DecisionGate = state.DecisionGate
	db.ProviderCandidates = taskDBProviderCandidates(state.ProviderCandidates)
	db.Diagnostics = taskDBDiagnostics(state.Diagnostics)
	sort.Slice(state.Tasks, func(i, j int) bool {
		return state.Tasks[i].ID < state.Tasks[j].ID
	})

	db = syncTaskDBRecords(db, state.Tasks, stamp)
	recountTransitions(&db)
	recountEvidence(&db)
	recountCommandReceipts(&db)
	return db
}
