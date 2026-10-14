package main

type Manifest struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	GeneratedDoc     string        `json:"generated_doc"`
	Workflow         string        `json:"workflow"`
	EvidenceArtifact string        `json:"evidence_artifact"`
	Sources          Sources       `json:"sources"`
	ApprovalEvent    ApprovalEvent `json:"approval_event"`
	TimeoutEvent     TimeoutEvent  `json:"timeout_event"`
	SourceChecks     []SourceCheck `json:"source_checks"`
	Assertions       []string      `json:"assertions"`
}

type Sources struct {
	SemanticActivityManifest string `json:"semantic_activity_manifest"`
	ProviderDraftManifest    string `json:"provider_draft_manifest"`
}

type ApprovalEvent struct {
	EventKind string `json:"event_kind"`
	EventType string `json:"event_type"`
}

type TimeoutEvent struct {
	EventKind     string `json:"event_kind"`
	ResultStatus  string `json:"result_status"`
	CancelCommand string `json:"cancel_command"`
}

type SourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}
