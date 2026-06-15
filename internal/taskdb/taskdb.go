// Package taskdb owns the public daemon's local riido-task-db.v1 persistence
// model and guarded mutation rules.
//
// It does not own workspace projection, mwsd synchronization, local IPC, or
// provider execution. Those contexts feed task rows into this package or consume
// its receipts through explicit adapters.
package taskdb

import (
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
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
