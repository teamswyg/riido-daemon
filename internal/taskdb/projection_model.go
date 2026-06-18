package taskdb

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
