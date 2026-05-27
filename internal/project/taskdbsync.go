package project

import (
	"sort"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
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

	records := make(map[string]taskdb.TaskRecord, len(db.Tasks)+len(state.Tasks))
	for _, record := range db.Tasks {
		records[record.ID] = record
	}
	sort.Slice(state.Tasks, func(i, j int) bool {
		return state.Tasks[i].ID < state.Tasks[j].ID
	})
	for _, sourceTask := range state.Tasks {
		record, exists := records[sourceTask.ID]
		if !exists {
			record = taskdb.TaskRecord{
				ID:        sourceTask.ID,
				State:     task.StateCreated,
				CreatedAt: stamp,
			}
			transition := taskdb.TaskTransitionRecord{
				ID:         initialTaskTransitionID(sourceTask.ID),
				TaskID:     sourceTask.ID,
				ToState:    task.StateCreated,
				EventType:  ir.EventTaskCreated,
				Actor:      "riido",
				Source:     "riido.mwsd.sync",
				Reason:     "document task source discovered",
				RecordedAt: stamp,
			}
			if !hasTransition(db.Transitions, transition.ID) {
				db.Transitions = append(db.Transitions, transition)
			}
		}
		record.ProjectID = sourceTask.ProjectID
		record.SourceDocumentID = sourceTask.SourceDocumentID
		record.SourceDocumentPath = sourceTask.SourceDocumentPath
		record.Title = sourceTask.Title
		record.Owner = sourceTask.Owner
		record.SourceStatus = sourceTask.SourceStatus
		record.RecommendedProvider = sourceTask.RecommendedProvider
		record.RecommendedDecisionLLM = sourceTask.RecommendedDecisionLLM
		record.RequiresHumanApproval = sourceTask.RequiresHumanApproval
		record.HarnessNextDirection = sourceTask.HarnessNextDirection
		record.UpdatedAt = stamp
		records[sourceTask.ID] = record
	}

	db.Tasks = mapToSortedTasks(records)
	recountTransitions(&db)
	recountEvidence(&db)
	recountCommandReceipts(&db)
	return db
}

func normalizeTaskDB(db taskdb.TaskDB) taskdb.TaskDB {
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
	sort.Slice(db.ProviderCandidates, func(i, j int) bool {
		return db.ProviderCandidates[i].ID < db.ProviderCandidates[j].ID
	})
	sort.Slice(db.Tasks, func(i, j int) bool {
		return db.Tasks[i].ID < db.Tasks[j].ID
	})
	sort.Slice(db.Transitions, func(i, j int) bool {
		return db.Transitions[i].RecordedAt < db.Transitions[j].RecordedAt ||
			(db.Transitions[i].RecordedAt == db.Transitions[j].RecordedAt && db.Transitions[i].ID < db.Transitions[j].ID)
	})
	sort.Slice(db.Evidence, func(i, j int) bool {
		return db.Evidence[i].RecordedAt < db.Evidence[j].RecordedAt ||
			(db.Evidence[i].RecordedAt == db.Evidence[j].RecordedAt && db.Evidence[i].ID < db.Evidence[j].ID)
	})
	sort.Slice(db.CommandReceipts, func(i, j int) bool {
		return db.CommandReceipts[i].RecordedAt < db.CommandReceipts[j].RecordedAt ||
			(db.CommandReceipts[i].RecordedAt == db.CommandReceipts[j].RecordedAt && db.CommandReceipts[i].ID < db.CommandReceipts[j].ID)
	})
	recountTransitions(&db)
	recountEvidence(&db)
	recountCommandReceipts(&db)
	return db
}

func taskDBProviderCandidates(candidates []ProviderCandidate) []taskdb.ProviderCandidate {
	out := make([]taskdb.ProviderCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		out = append(out, taskdb.ProviderCandidate{
			ID:               candidate.ID,
			SourceWorkflow:   candidate.SourceWorkflow,
			Available:        candidate.Available,
			ApprovalRequired: candidate.ApprovalRequired,
		})
	}
	return out
}

func taskDBDiagnostics(diagnostics []ProjectionDiagnostic) []taskdb.ProjectionDiagnostic {
	out := make([]taskdb.ProjectionDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		out = append(out, taskdb.ProjectionDiagnostic{
			Severity: diagnostic.Severity,
			Code:     diagnostic.Code,
			Message:  diagnostic.Message,
		})
	}
	return out
}

func mapToSortedTasks(records map[string]taskdb.TaskRecord) []taskdb.TaskRecord {
	tasks := make([]taskdb.TaskRecord, 0, len(records))
	for _, record := range records {
		tasks = append(tasks, record)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	return tasks
}

func recountTransitions(db *taskdb.TaskDB) {
	counts := make(map[string]int, len(db.Tasks))
	for _, transition := range db.Transitions {
		counts[transition.TaskID]++
	}
	for index := range db.Tasks {
		db.Tasks[index].TransitionCount = counts[db.Tasks[index].ID]
	}
}

func recountEvidence(db *taskdb.TaskDB) {
	counts := make(map[string]int, len(db.Tasks))
	for _, evidence := range db.Evidence {
		counts[evidence.TaskID]++
	}
	for index := range db.Tasks {
		db.Tasks[index].EvidenceCount = counts[db.Tasks[index].ID]
	}
}

func recountCommandReceipts(db *taskdb.TaskDB) {
	counts := make(map[string]int, len(db.Tasks))
	for _, receipt := range db.CommandReceipts {
		counts[receipt.TaskID]++
	}
	for index := range db.Tasks {
		db.Tasks[index].CommandReceiptCount = counts[db.Tasks[index].ID]
	}
}

func hasTransition(transitions []taskdb.TaskTransitionRecord, id string) bool {
	for _, transition := range transitions {
		if transition.ID == id {
			return true
		}
	}
	return false
}

func initialTaskTransitionID(taskID string) string {
	return "transition:" + taskID + ":created"
}

func timestamp(now time.Time) string {
	return now.UTC().Format(time.RFC3339Nano)
}
