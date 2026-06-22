package main

type providerRunEvidence struct {
	Status    string                `json:"status"`
	Providers []providerRunProvider `json:"providers"`
}

type providerRunProvider struct {
	ID     string     `json:"id"`
	Repair *runRepair `json:"repair,omitempty"`
}

type runRepair struct {
	ProviderID       string `json:"provider_id,omitempty"`
	Class            string `json:"class"`
	Owner            string `json:"owner"`
	Mode             string `json:"mode"`
	Summary          string `json:"summary"`
	SuggestedCommand string `json:"suggested_command,omitempty"`
}
