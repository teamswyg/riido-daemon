package main

type Manifest struct {
	SchemaVersion    string         `json:"schema_version"`
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	GeneratedDoc     string         `json:"generated_doc"`
	Workflow         string         `json:"workflow"`
	EvidenceArtifact string         `json:"evidence_artifact"`
	Source           string         `json:"source"`
	MappedEvents     []MappedEvent  `json:"mapped_events"`
	SkippedEvents    []SkippedEvent `json:"skipped_events"`
	Assertions       []string       `json:"assertions"`
}

type MappedEvent struct {
	EventKind      string `json:"event_kind"`
	EventKindConst string `json:"event_kind_const"`
	EventTypeConst string `json:"event_type_const"`
	EventType      string `json:"event_type"`
}

type SkippedEvent struct {
	EventKind      string `json:"event_kind"`
	EventKindConst string `json:"event_kind_const"`
	Reason         string `json:"reason"`
}
