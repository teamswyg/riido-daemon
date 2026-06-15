// Package taskdb owns the public daemon's local riido-task-db.v1 persistence
// model and guarded mutation rules.
//
// It does not own workspace projection, mwsd synchronization, local IPC, or
// provider execution. Those contexts feed task rows into this package or consume
// its receipts through explicit adapters.
package taskdb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

const (
	TaskDBSchemaVersion       = "riido-task-db.v1"
	TaskCommandReplayPolicyV1 = "command-id-idempotent-replay.v1"
	TaskEvidenceValidationV1  = "deterministic-command-exit-code.v1"
)

// TaskDB is Riido's first local task transition database.
//
// It stays dependency-light on purpose: the local daemon source is single host
// and local-only, so atomic JSON replacement gives us a simple, inspectable
// database before we introduce a heavier embedded store.
type TaskDB struct {
	SchemaVersion          string                     `json:"schema_version"`
	ProjectionVersion      string                     `json:"projection_version"`
	Root                   string                     `json:"root"`
	Domain                 string                     `json:"domain"`
	UpdatedAt              string                     `json:"updated_at"`
	RecommendedProvider    string                     `json:"recommended_provider"`
	RecommendedDecisionLLM string                     `json:"recommended_decision_llm"`
	DecisionGate           string                     `json:"decision_gate"`
	ProviderCandidates     []ProviderCandidate        `json:"provider_candidates"`
	Tasks                  []TaskRecord               `json:"tasks"`
	Transitions            []TaskTransitionRecord     `json:"transitions"`
	Evidence               []TaskEvidenceRecord       `json:"evidence"`
	CommandReceipts        []TaskCommandReceiptRecord `json:"command_receipts"`
	Diagnostics            []ProjectionDiagnostic     `json:"diagnostics"`
}

type ProviderCandidate struct {
	ID               string `json:"id"`
	SourceWorkflow   string `json:"source_workflow"`
	Available        bool   `json:"available"`
	ApprovalRequired bool   `json:"approval_required"`
}

type ProjectionDiagnostic struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type TaskRecord struct {
	ID                     string         `json:"id"`
	ProjectID              string         `json:"project_id"`
	State                  task.TaskState `json:"state"`
	SourceDocumentID       string         `json:"source_document_id"`
	SourceDocumentPath     string         `json:"source_document_path"`
	Title                  string         `json:"title"`
	Owner                  string         `json:"owner"`
	SourceStatus           string         `json:"source_status"`
	RecommendedProvider    string         `json:"recommended_provider"`
	RecommendedDecisionLLM string         `json:"recommended_decision_llm"`
	RequiresHumanApproval  bool           `json:"requires_human_approval"`
	HarnessNextDirection   string         `json:"harness_next_direction"`
	CreatedAt              string         `json:"created_at"`
	UpdatedAt              string         `json:"updated_at"`
	TransitionCount        int            `json:"transition_count"`
	EvidenceCount          int            `json:"evidence_count"`
	CommandReceiptCount    int            `json:"command_receipt_count"`
}

type TaskTransitionRecord struct {
	ID               string         `json:"id"`
	TaskID           string         `json:"task_id"`
	FromState        task.TaskState `json:"from_state"`
	ToState          task.TaskState `json:"to_state"`
	EventType        ir.EventType   `json:"event_type"`
	Actor            string         `json:"actor"`
	Source           string         `json:"source"`
	Reason           string         `json:"reason"`
	CommandReceiptID string         `json:"command_receipt_id"`
	RecordedAt       string         `json:"recorded_at"`
}

type TaskEvidenceRecord struct {
	ID                string `json:"id"`
	TaskID            string `json:"task_id"`
	ProjectID         string `json:"project_id"`
	DocumentID        string `json:"document_id"`
	DocumentPath      string `json:"document_path"`
	Command           string `json:"command"`
	ExitCode          int    `json:"exit_code"`
	Result            string `json:"result"`
	ValidationGate    string `json:"validation_gate"`
	ProviderRunID     string `json:"provider_run_id"`
	ProviderRunResult string `json:"provider_run_result"`
	Actor             string `json:"actor"`
	Source            string `json:"source"`
	Summary           string `json:"summary"`
	CommandReceiptID  string `json:"command_receipt_id"`
	RecordedAt        string `json:"recorded_at"`
}

type TaskCommandReceiptRecord struct {
	ID                     string `json:"id"`
	CommandID              string `json:"command_id"`
	Kind                   string `json:"kind"`
	TaskID                 string `json:"task_id"`
	Actor                  string `json:"actor"`
	Source                 string `json:"source"`
	Provider               string `json:"provider"`
	DecisionLLM            string `json:"decision_llm"`
	ApprovalID             string `json:"approval_id"`
	DecisionGate           string `json:"decision_gate"`
	RequiresHumanApproval  bool   `json:"requires_human_approval"`
	RecommendedProvider    string `json:"recommended_provider"`
	RecommendedDecisionLLM string `json:"recommended_decision_llm"`
	HarnessNextDirection   string `json:"harness_next_direction"`
	GuardDecision          string `json:"guard_decision"`
	GuardReason            string `json:"guard_reason"`
	ReplayPolicy           string `json:"replay_policy"`
	TransitionID           string `json:"transition_id,omitempty"`
	EvidenceID             string `json:"evidence_id,omitempty"`
	Result                 string `json:"result"`
	RecordedAt             string `json:"recorded_at"`
}

type TaskMutationGuardInput struct {
	CommandID   string
	Provider    string
	DecisionLLM string
	ApprovalID  string
}

type TaskTransitionInput struct {
	TaskID  string
	ToState task.TaskState
	Event   ir.EventType
	Actor   string
	Source  string
	Reason  string
	Guard   TaskMutationGuardInput
}

type TaskEvidenceInput struct {
	TaskID            string
	Command           string
	ExitCode          int
	Result            string
	Actor             string
	Source            string
	Summary           string
	ValidationGate    string
	ProviderRunID     string
	ProviderRunResult string
	Guard             TaskMutationGuardInput
}

func EmptyTaskDB() TaskDB {
	return TaskDB{
		SchemaVersion:   TaskDBSchemaVersion,
		Tasks:           []TaskRecord{},
		Transitions:     []TaskTransitionRecord{},
		Evidence:        []TaskEvidenceRecord{},
		CommandReceipts: []TaskCommandReceiptRecord{},
		Diagnostics:     []ProjectionDiagnostic{},
	}
}

func ApplyTaskTransition(existing TaskDB, taskID string, to task.TaskState, event ir.EventType, actor, source, reason string, now time.Time) (TaskDB, TaskTransitionRecord, error) {
	updated, transition, _, err := ApplyGuardedTaskTransition(existing, TaskTransitionInput{
		TaskID:  taskID,
		ToState: to,
		Event:   event,
		Actor:   actor,
		Source:  source,
		Reason:  reason,
		Guard: TaskMutationGuardInput{
			ApprovalID: "approval.riido.legacy",
		},
	}, now)
	return updated, transition, err
}

func ApplyGuardedTaskTransition(existing TaskDB, input TaskTransitionInput, now time.Time) (TaskDB, TaskTransitionRecord, TaskCommandReceiptRecord, error) {
	if input.TaskID == "" {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBInput, "transition.validate", "task id is empty")
	}
	if !isKnownTaskState(input.ToState) {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBState, "transition.validate", "unknown target state: %s", input.ToState)
	}
	if !input.Event.IsTransition() {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBState, "transition.validate", "event %q is not a transition event", input.Event)
	}
	db := normalizeTaskDB(existing)
	index := -1
	for i, record := range db.Tasks {
		if record.ID == input.TaskID {
			index = i
			break
		}
	}
	if index < 0 {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBState, "transition.find-task", "task %s not found", input.TaskID)
	}
	actor := textutil.FirstNonEmpty(input.Actor, "human")
	source := textutil.FirstNonEmpty(input.Source, "riido-cli")
	replayedTransition, replayedReceipt, replayed, err := replayExistingTaskTransition(db, input, actor, source)
	if err != nil {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, err
	}
	if replayed {
		return db, replayedTransition, replayedReceipt, nil
	}
	from := db.Tasks[index].State
	if !task.ValidateTransition(from, input.ToState, input.Event) {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBState, "transition.apply", "illegal task transition: %s --%s--> %s", from, input.Event, input.ToState)
	}
	stamp := timestamp(now)
	receipt, err := buildTaskCommandReceipt(db, db.Tasks[index], "transition", actor, source, input.Guard, now, len(db.CommandReceipts)+1)
	if err != nil {
		return TaskDB{}, TaskTransitionRecord{}, TaskCommandReceiptRecord{}, err
	}
	transition := TaskTransitionRecord{
		ID:               transitionID(input.TaskID, input.Event, now, len(db.Transitions)+1),
		TaskID:           input.TaskID,
		FromState:        from,
		ToState:          input.ToState,
		EventType:        input.Event,
		Actor:            actor,
		Source:           source,
		Reason:           input.Reason,
		CommandReceiptID: receipt.ID,
		RecordedAt:       stamp,
	}
	receipt.TransitionID = transition.ID
	db.Transitions = append(db.Transitions, transition)
	db.CommandReceipts = append(db.CommandReceipts, receipt)
	db.Tasks[index].State = input.ToState
	db.Tasks[index].UpdatedAt = stamp
	db.UpdatedAt = stamp
	recountTransitions(&db)
	recountEvidence(&db)
	recountCommandReceipts(&db)
	sort.Slice(db.Tasks, func(i, j int) bool {
		return db.Tasks[i].ID < db.Tasks[j].ID
	})
	return db, transition, receipt, nil
}

func AddTaskEvidence(existing TaskDB, input TaskEvidenceInput, now time.Time) (TaskDB, TaskEvidenceRecord, error) {
	updated, evidence, _, err := AddGuardedTaskEvidence(existing, input, now)
	return updated, evidence, err
}

func AddGuardedTaskEvidence(existing TaskDB, input TaskEvidenceInput, now time.Time) (TaskDB, TaskEvidenceRecord, TaskCommandReceiptRecord, error) {
	if input.TaskID == "" {
		return TaskDB{}, TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBInput, "evidence.validate", "task id is empty")
	}
	if strings.TrimSpace(input.Command) == "" {
		return TaskDB{}, TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBInput, "evidence.validate", "evidence command is empty")
	}
	db := normalizeTaskDB(existing)
	index := -1
	for i, record := range db.Tasks {
		if record.ID == input.TaskID {
			index = i
			break
		}
	}
	if index < 0 {
		return TaskDB{}, TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBState, "evidence.find-task", "task %s not found", input.TaskID)
	}
	stamp := timestamp(now)
	taskRecord := db.Tasks[index]
	actor := textutil.FirstNonEmpty(input.Actor, "daemon")
	source := textutil.FirstNonEmpty(input.Source, "riido-cli")
	replayedEvidence, replayedReceipt, replayed, err := replayExistingTaskEvidence(db, input, actor, source)
	if err != nil {
		return TaskDB{}, TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, err
	}
	if replayed {
		return db, replayedEvidence, replayedReceipt, nil
	}
	receipt, err := buildTaskCommandReceipt(db, taskRecord, "evidence", actor, source, input.Guard, now, len(db.CommandReceipts)+1)
	if err != nil {
		return TaskDB{}, TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, err
	}
	evidenceResult := normalizeEvidenceResult(input.Result, input.ExitCode)
	validationGate := textutil.FirstNonEmpty(input.ValidationGate, TaskEvidenceValidationV1)
	providerRunID := textutil.FirstNonEmpty(input.ProviderRunID, receipt.CommandID)
	providerRunResult := textutil.FirstNonEmpty(input.ProviderRunResult, evidenceResult)
	evidence := TaskEvidenceRecord{
		ID:                evidenceID(input.TaskID, now, len(db.Evidence)+1),
		TaskID:            input.TaskID,
		ProjectID:         taskRecord.ProjectID,
		DocumentID:        taskRecord.SourceDocumentID,
		DocumentPath:      taskRecord.SourceDocumentPath,
		Command:           strings.TrimSpace(input.Command),
		ExitCode:          input.ExitCode,
		Result:            evidenceResult,
		ValidationGate:    validationGate,
		ProviderRunID:     providerRunID,
		ProviderRunResult: providerRunResult,
		Actor:             actor,
		Source:            source,
		Summary:           input.Summary,
		CommandReceiptID:  receipt.ID,
		RecordedAt:        stamp,
	}
	receipt.EvidenceID = evidence.ID
	receipt.Result = evidence.Result
	db.Evidence = append(db.Evidence, evidence)
	db.CommandReceipts = append(db.CommandReceipts, receipt)
	db.Tasks[index].UpdatedAt = stamp
	db.UpdatedAt = stamp
	recountTransitions(&db)
	recountEvidence(&db)
	recountCommandReceipts(&db)
	sort.Slice(db.Tasks, func(i, j int) bool {
		return db.Tasks[i].ID < db.Tasks[j].ID
	})
	return db, evidence, receipt, nil
}

func ParseTaskState(value string) (task.TaskState, error) {
	candidate := task.TaskState(value)
	if !isKnownTaskState(candidate) {
		return "", taskDBErrorf(ErrTaskDBState, "parse-state", "unknown task state: %s", value)
	}
	return candidate, nil
}

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

func replayExistingTaskTransition(db TaskDB, input TaskTransitionInput, actor string, source string) (TaskTransitionRecord, TaskCommandReceiptRecord, bool, error) {
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

func replayExistingTaskEvidence(db TaskDB, input TaskEvidenceInput, actor string, source string) (TaskEvidenceRecord, TaskCommandReceiptRecord, bool, error) {
	receipt, found, err := findCommandReceiptByCommandID(db, input.Guard.CommandID)
	if err != nil || !found {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, found, err
	}
	if err := validateCommandReceiptReplay(receipt, "evidence", input.TaskID, actor, source, input.Guard); err != nil {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, err
	}
	if receipt.EvidenceID == "" {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, taskDBErrorf(ErrTaskDBReplay, "evidence.replay", "command_id %s replay cannot find linked evidence id", receipt.CommandID)
	}
	evidence, ok := findEvidenceByID(db.Evidence, receipt.EvidenceID)
	if !ok {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, taskDBErrorf(ErrTaskDBReplay, "evidence.replay", "command_id %s replay cannot find evidence %s", receipt.CommandID, receipt.EvidenceID)
	}
	if evidence.TaskID != input.TaskID {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "task_id")
	}
	if evidence.Command != strings.TrimSpace(input.Command) {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "command")
	}
	if evidence.ExitCode != input.ExitCode {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "exit_code")
	}
	if evidence.Result != normalizeEvidenceResult(input.Result, input.ExitCode) {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "result")
	}
	validationGate := textutil.FirstNonEmpty(input.ValidationGate, TaskEvidenceValidationV1)
	providerRunID := textutil.FirstNonEmpty(input.ProviderRunID, receipt.CommandID)
	providerRunResult := textutil.FirstNonEmpty(input.ProviderRunResult, evidence.Result)
	if !replayStringFieldMatches(evidence.ValidationGate, validationGate, input.ValidationGate != "") {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "validation_gate")
	}
	if !replayStringFieldMatches(evidence.ProviderRunID, providerRunID, input.ProviderRunID != "") {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "provider_run_id")
	}
	if !replayStringFieldMatches(evidence.ProviderRunResult, providerRunResult, input.ProviderRunResult != "") {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "provider_run_result")
	}
	if evidence.Actor != actor {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "actor")
	}
	if evidence.Source != source {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "source")
	}
	if evidence.Summary != input.Summary {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, commandReplayMismatch(receipt.CommandID, "summary")
	}
	return evidence, receipt, true, nil
}

func validateCommandReceiptReplay(receipt TaskCommandReceiptRecord, kind string, taskID string, actor string, source string, guard TaskMutationGuardInput) error {
	if receipt.Kind != kind {
		return commandReplayMismatch(receipt.CommandID, "kind")
	}
	if receipt.TaskID != taskID {
		return commandReplayMismatch(receipt.CommandID, "task_id")
	}
	if receipt.Actor != actor {
		return commandReplayMismatch(receipt.CommandID, "actor")
	}
	if receipt.Source != source {
		return commandReplayMismatch(receipt.CommandID, "source")
	}
	if receipt.ApprovalID != strings.TrimSpace(guard.ApprovalID) {
		return commandReplayMismatch(receipt.CommandID, "approval_id")
	}
	if strings.TrimSpace(guard.Provider) != "" && receipt.Provider != strings.TrimSpace(guard.Provider) {
		return commandReplayMismatch(receipt.CommandID, "provider")
	}
	if strings.TrimSpace(guard.DecisionLLM) != "" && receipt.DecisionLLM != strings.TrimSpace(guard.DecisionLLM) {
		return commandReplayMismatch(receipt.CommandID, "decision_llm")
	}
	if receipt.GuardDecision != "accepted" {
		return taskDBErrorf(ErrTaskDBReplay, "receipt.replay", "command_id %s replay cannot reuse receipt with guard decision %s", receipt.CommandID, receipt.GuardDecision)
	}
	return nil
}

func findCommandReceiptByCommandID(db TaskDB, commandID string) (TaskCommandReceiptRecord, bool, error) {
	commandID = strings.TrimSpace(commandID)
	if commandID == "" {
		return TaskCommandReceiptRecord{}, false, nil
	}
	var found TaskCommandReceiptRecord
	hasFound := false
	for _, receipt := range db.CommandReceipts {
		if receipt.CommandID != commandID {
			continue
		}
		if hasFound {
			return TaskCommandReceiptRecord{}, true, taskDBErrorf(ErrTaskDBReplay, "receipt.find-by-command-id", "command_id %s is not unique in task DB", commandID)
		}
		found = receipt
		hasFound = true
	}
	return found, hasFound, nil
}

func findTransitionByID(transitions []TaskTransitionRecord, id string) (TaskTransitionRecord, bool) {
	for _, transition := range transitions {
		if transition.ID == id {
			return transition, true
		}
	}
	return TaskTransitionRecord{}, false
}

func findEvidenceByID(evidenceRecords []TaskEvidenceRecord, id string) (TaskEvidenceRecord, bool) {
	for _, evidence := range evidenceRecords {
		if evidence.ID == id {
			return evidence, true
		}
	}
	return TaskEvidenceRecord{}, false
}

func commandReplayMismatch(commandID string, field string) error {
	return taskDBErrorf(ErrTaskDBReplay, "receipt.replay", "command_id %s replay mismatch on %s", commandID, field)
}

func replayStringFieldMatches(existing string, expected string, required bool) bool {
	if existing == expected {
		return true
	}
	return existing == "" && !required
}

func buildTaskCommandReceipt(db TaskDB, taskRecord TaskRecord, kind string, actor string, source string, guard TaskMutationGuardInput, now time.Time, ordinal int) (TaskCommandReceiptRecord, error) {
	provider := textutil.FirstNonEmpty(guard.Provider, taskRecord.RecommendedProvider)
	provider = textutil.FirstNonEmpty(provider, db.RecommendedProvider)
	decisionLLM := textutil.FirstNonEmpty(guard.DecisionLLM, taskRecord.RecommendedDecisionLLM)
	decisionLLM = textutil.FirstNonEmpty(decisionLLM, db.RecommendedDecisionLLM)
	approvalID := strings.TrimSpace(guard.ApprovalID)
	commandID := strings.TrimSpace(guard.CommandID)
	if commandID == "" {
		commandID = generatedCommandID(taskRecord.ID, kind, now, ordinal)
	}
	requiresHumanApproval := taskRecord.RequiresHumanApproval || db.DecisionGate == "human-approval-required"
	receipt := TaskCommandReceiptRecord{
		ID:                     commandReceiptID(taskRecord.ID, kind, now, ordinal),
		CommandID:              commandID,
		Kind:                   kind,
		TaskID:                 taskRecord.ID,
		Actor:                  actor,
		Source:                 source,
		Provider:               provider,
		DecisionLLM:            decisionLLM,
		ApprovalID:             approvalID,
		DecisionGate:           db.DecisionGate,
		RequiresHumanApproval:  requiresHumanApproval,
		RecommendedProvider:    textutil.FirstNonEmpty(taskRecord.RecommendedProvider, db.RecommendedProvider),
		RecommendedDecisionLLM: textutil.FirstNonEmpty(taskRecord.RecommendedDecisionLLM, db.RecommendedDecisionLLM),
		HarnessNextDirection:   taskRecord.HarnessNextDirection,
		ReplayPolicy:           TaskCommandReplayPolicyV1,
		RecordedAt:             timestamp(now),
	}
	if requiresHumanApproval && approvalID == "" {
		return TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBGuard, "receipt.build", "task %s requires approval_id before %s mutation", taskRecord.ID, kind)
	}
	if provider == "" {
		return TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBGuard, "receipt.build", "task %s has no provider for %s mutation", taskRecord.ID, kind)
	}
	if !providerCandidateAvailable(db.ProviderCandidates, provider) {
		return TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBGuard, "receipt.build", "provider %s is not an available orchestration candidate for task %s", provider, taskRecord.ID)
	}
	if receipt.RecommendedDecisionLLM != "" && decisionLLM != receipt.RecommendedDecisionLLM {
		return TaskCommandReceiptRecord{}, taskDBErrorf(ErrTaskDBGuard, "receipt.build", "decision LLM %s does not match recommended decision LLM %s for task %s", decisionLLM, receipt.RecommendedDecisionLLM, taskRecord.ID)
	}
	receipt.GuardDecision = "accepted"
	receipt.GuardReason = "approval_id and orchestration provider candidate accepted"
	receipt.Result = "accepted"
	return receipt, nil
}

func providerCandidateAvailable(candidates []ProviderCandidate, provider string) bool {
	if len(candidates) == 0 {
		return true
	}
	for _, candidate := range candidates {
		if candidate.ID == provider {
			return candidate.Available
		}
	}
	return false
}

func transitionID(taskID string, event ir.EventType, now time.Time, ordinal int) string {
	return fmt.Sprintf("transition:%s:%s:%s:%04d", taskID, event, now.UTC().Format("20060102T150405.000000000Z"), ordinal)
}

func evidenceID(taskID string, now time.Time, ordinal int) string {
	return fmt.Sprintf("evidence:%s:%s:%04d", taskID, now.UTC().Format("20060102T150405.000000000Z"), ordinal)
}

func commandReceiptID(taskID string, kind string, now time.Time, ordinal int) string {
	return fmt.Sprintf("receipt:%s:%s:%s:%04d", kind, taskID, now.UTC().Format("20060102T150405.000000000Z"), ordinal)
}

func generatedCommandID(taskID string, kind string, now time.Time, ordinal int) string {
	return fmt.Sprintf("command:%s:%s:%s:%04d", kind, taskID, now.UTC().Format("20060102T150405.000000000Z"), ordinal)
}

func timestamp(now time.Time) string {
	return now.UTC().Format(time.RFC3339Nano)
}

func isKnownTaskState(value task.TaskState) bool {
	for _, state := range task.AllStates() {
		if state == value {
			return true
		}
	}
	return false
}

func normalizeEvidenceResult(result string, exitCode int) string {
	switch strings.ToLower(strings.TrimSpace(result)) {
	case "passed", "failed", "unknown":
		return strings.ToLower(strings.TrimSpace(result))
	case "":
		if exitCode == 0 {
			return "passed"
		}
		return "failed"
	default:
		return "unknown"
	}
}
