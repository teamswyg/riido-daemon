package riidoapi

import (
	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/validation"
)

type ValidateRequest struct {
	TaskID         string `json:"task_id"`
	Command        string `json:"command"`
	Workdir        string `json:"workdir"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	Actor          string `json:"actor"`
	Source         string `json:"source"`
	Summary        string `json:"summary"`
	Provider       string `json:"provider"`
	DecisionLLM    string `json:"decision_llm"`
	ApprovalID     string `json:"approval_id"`
	CommandID      string `json:"command_id"`
	ValidationGate string `json:"validation_gate"`
}

type ValidateResponse struct {
	TaskDBPath        string                           `json:"task_db_path"`
	Task              taskdb.TaskRecord                `json:"task"`
	Validation        validation.CommandResult         `json:"validation"`
	Evidence          taskdb.TaskEvidenceRecord        `json:"evidence"`
	Receipt           taskdb.TaskCommandReceiptRecord  `json:"receipt"`
	Transition        *taskdb.TaskTransitionRecord     `json:"transition,omitempty"`
	TransitionReceipt *taskdb.TaskCommandReceiptRecord `json:"transition_receipt,omitempty"`
}
