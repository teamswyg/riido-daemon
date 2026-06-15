package taskdb

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
)

func DefaultTaskDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "riido", "task-db.json"), nil
}

func LoadTaskDBOrEmpty(path string) (TaskDB, error) {
	db, err := LoadTaskDB(path)
	if os.IsNotExist(err) {
		return EmptyTaskDB(), nil
	}
	return db, err
}

func SaveTaskDB(path string, db TaskDB) error {
	if path == "" {
		return taskDBErrorf(ErrTaskDBInput, "save", "task DB path is empty")
	}
	if err := fileutil.WriteJSONAtomic(path, normalizeTaskDB(db)); err != nil {
		return taskDBWrapf(ErrTaskDBPersistence, "save", err, "save task DB")
	}
	return nil
}

func LoadTaskDB(path string) (TaskDB, error) {
	var db TaskDB
	data, err := os.ReadFile(path)
	if err != nil {
		return db, err
	}
	if err := json.Unmarshal(data, &db); err != nil {
		return db, taskDBWrapf(ErrTaskDBPersistence, "load.decode", err, "decode task DB")
	}
	if db.SchemaVersion != TaskDBSchemaVersion {
		return db, taskDBErrorf(ErrTaskDBSchema, "load.validate-schema", "task DB schema mismatch: got %q want %q", db.SchemaVersion, TaskDBSchemaVersion)
	}
	return normalizeTaskDB(db), nil
}

func normalizeTaskDB(db TaskDB) TaskDB {
	if db.SchemaVersion == "" {
		db.SchemaVersion = TaskDBSchemaVersion
	}
	if db.Tasks == nil {
		db.Tasks = []TaskRecord{}
	}
	if db.Transitions == nil {
		db.Transitions = []TaskTransitionRecord{}
	}
	if db.Evidence == nil {
		db.Evidence = []TaskEvidenceRecord{}
	}
	if db.CommandReceipts == nil {
		db.CommandReceipts = []TaskCommandReceiptRecord{}
	}
	if db.Diagnostics == nil {
		db.Diagnostics = []ProjectionDiagnostic{}
	}
	if db.ProviderCandidates == nil {
		db.ProviderCandidates = []ProviderCandidate{}
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

func recountTransitions(db *TaskDB) {
	counts := make(map[string]int, len(db.Tasks))
	for _, transition := range db.Transitions {
		counts[transition.TaskID]++
	}
	for index := range db.Tasks {
		db.Tasks[index].TransitionCount = counts[db.Tasks[index].ID]
	}
}

func recountEvidence(db *TaskDB) {
	counts := make(map[string]int, len(db.Tasks))
	for _, evidence := range db.Evidence {
		counts[evidence.TaskID]++
	}
	for index := range db.Tasks {
		db.Tasks[index].EvidenceCount = counts[db.Tasks[index].ID]
	}
}

func recountCommandReceipts(db *TaskDB) {
	counts := make(map[string]int, len(db.Tasks))
	for _, receipt := range db.CommandReceipts {
		counts[receipt.TaskID]++
	}
	for index := range db.Tasks {
		db.Tasks[index].CommandReceiptCount = counts[db.Tasks[index].ID]
	}
}

func replayExistingTaskTransition(db TaskDB, input TaskTransitionInput, actor, source string) (TaskTransitionRecord, TaskCommandReceiptRecord, bool, error) {
	receipt, found, err := findCommandReceiptByCommandID(db, input.Guard.CommandID)
	if err != nil || !found {
		return TaskTransitionRecord{}, TaskCommandReceiptRecord{}, found, err
	}
	if err := validateCommandReceiptReplay(receipt, "transition", input.TaskID, actor, source, input.Guard); err != nil {
		return TaskTransitionRecord{}, TaskCommandReceiptRecord{}, true, err
	}
	if receipt.TransitionID == "" {
		return TaskTransitionRecord{}, TaskCommandReceiptRecord{}, true, taskDBErrorf(ErrTaskDBReplay, "transition.replay", "command_id %s replay cannot find linked transition id", receipt.CommandID)
	}
	transition, ok := findTransitionByID(db.Transitions, receipt.TransitionID)
	if !ok {
		return TaskTransitionRecord{}, TaskCommandReceiptRecord{}, true, taskDBErrorf(ErrTaskDBReplay, "transition.replay", "command_id %s replay cannot find transition %s", receipt.CommandID, receipt.TransitionID)
	}
	if transition.TaskID != input.TaskID {
		return TaskTransitionRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "task_id")
	}
	if transition.ToState != input.ToState {
		return TaskTransitionRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "to_state")
	}
	if transition.EventType != input.Event {
		return TaskTransitionRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "event_type")
	}
	if transition.Actor != actor {
		return TaskTransitionRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "actor")
	}
	if transition.Source != source {
		return TaskTransitionRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "source")
	}
	if transition.Reason != input.Reason {
		return TaskTransitionRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "reason")
	}
	return transition, receipt, true, nil
}
