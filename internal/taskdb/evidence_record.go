package taskdb

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
