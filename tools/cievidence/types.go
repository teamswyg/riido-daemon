package main

type options struct {
	Workflow    string
	ID          string
	EvidenceOut string
}

type evidence struct {
	SchemaVersion string     `json:"schema_version"`
	ID            string     `json:"id"`
	Status        string     `json:"status"`
	Workflow      string     `json:"workflow"`
	Required      []required `json:"required_commands"`
	Problems      []string   `json:"problem_summaries,omitempty"`
}

type required struct {
	Command string `json:"command"`
	Found   bool   `json:"found"`
}
