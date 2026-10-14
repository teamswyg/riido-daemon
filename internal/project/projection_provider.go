package project

type ProviderCandidate struct {
	ID               string `json:"id"`
	SourceWorkflow   string `json:"source_workflow"`
	Available        bool   `json:"available"`
	ApprovalRequired bool   `json:"approval_required"`
}

type ProviderRunSummary struct {
	ID        string `json:"id"`
	Direction string `json:"direction"`
	Source    string `json:"source"`
	Provider  string `json:"provider"`
	Command   string `json:"command"`
	Result    string `json:"result"`
}
