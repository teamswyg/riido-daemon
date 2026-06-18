package taskdb

import "github.com/teamswyg/riido-contracts/task"

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
