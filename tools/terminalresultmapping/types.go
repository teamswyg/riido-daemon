package main

type Manifest struct {
	SchemaVersion    string          `json:"schema_version"`
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	GeneratedDoc     string          `json:"generated_doc"`
	Workflow         string          `json:"workflow"`
	EvidenceArtifact string          `json:"evidence_artifact"`
	Sources          Sources         `json:"sources"`
	Mappings         []StatusMapping `json:"mappings"`
	Defaults         Defaults        `json:"defaults"`
	Assertions       []string        `json:"assertions"`
}

type Sources struct {
	ResultStatus   string `json:"result_status"`
	TerminalResult string `json:"terminal_result"`
}

type StatusMapping struct {
	Status         string `json:"status"`
	StatusConst    string `json:"status_const"`
	EventTypeConst string `json:"event_type_const"`
	EventType      string `json:"event_type"`
}

type Defaults struct {
	EmptyStatusConst       string `json:"empty_status_const"`
	FallbackEventTypeConst string `json:"fallback_event_type_const"`
	FallbackEventType      string `json:"fallback_event_type"`
}
