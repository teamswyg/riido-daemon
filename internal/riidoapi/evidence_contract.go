package riidoapi

import "github.com/teamswyg/riido-daemon/internal/taskdb"

type EvidenceRequest struct {
	TaskID            string `json:"task_id"`
	Command           string `json:"command"`
	ExitCode          int    `json:"exit_code"`
	Result            string `json:"result"`
	Actor             string `json:"actor"`
	Source            string `json:"source"`
	Summary           string `json:"summary"`
	Provider          string `json:"provider"`
	DecisionLLM       string `json:"decision_llm"`
	ApprovalID        string `json:"approval_id"`
	CommandID         string `json:"command_id"`
	ValidationGate    string `json:"validation_gate"`
	ProviderRunID     string `json:"provider_run_id"`
	ProviderRunResult string `json:"provider_run_result"`
}

type EvidenceResponse struct {
	TaskDBPath string                          `json:"task_db_path"`
	Task       taskdb.TaskRecord               `json:"task"`
	Evidence   taskdb.TaskEvidenceRecord       `json:"evidence"`
	Receipt    taskdb.TaskCommandReceiptRecord `json:"receipt"`
}
