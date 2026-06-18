package project

import (
	"sort"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func normalizeTaskDB(db taskdb.TaskDB) taskdb.TaskDB {
	db = ensureTaskDBSlices(db)
	sort.Slice(db.ProviderCandidates, func(i, j int) bool {
		return db.ProviderCandidates[i].ID < db.ProviderCandidates[j].ID
	})
	sort.Slice(db.Tasks, func(i, j int) bool {
		return db.Tasks[i].ID < db.Tasks[j].ID
	})
	sortTaskDBTransitions(db.Transitions)
	sortTaskDBEvidence(db.Evidence)
	sortTaskDBCommandReceipts(db.CommandReceipts)
	recountTransitions(&db)
	recountEvidence(&db)
	recountCommandReceipts(&db)
	return db
}

func ensureTaskDBSlices(db taskdb.TaskDB) taskdb.TaskDB {
	if db.SchemaVersion == "" {
		db.SchemaVersion = taskdb.TaskDBSchemaVersion
	}
	if db.Tasks == nil {
		db.Tasks = []taskdb.TaskRecord{}
	}
	if db.Transitions == nil {
		db.Transitions = []taskdb.TaskTransitionRecord{}
	}
	if db.Evidence == nil {
		db.Evidence = []taskdb.TaskEvidenceRecord{}
	}
	if db.CommandReceipts == nil {
		db.CommandReceipts = []taskdb.TaskCommandReceiptRecord{}
	}
	if db.Diagnostics == nil {
		db.Diagnostics = []taskdb.ProjectionDiagnostic{}
	}
	if db.ProviderCandidates == nil {
		db.ProviderCandidates = []taskdb.ProviderCandidate{}
	}
	return db
}
