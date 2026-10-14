package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type manifest struct {
	SchemaVersion        string           `json:"schema_version"`
	LoopSource           string           `json:"loop_source,omitempty"`
	ID                   string           `json:"id"`
	Title                string           `json:"title"`
	GeneratedDoc         string           `json:"generated_doc"`
	Workflow             string           `json:"workflow"`
	EvidenceArtifact     string           `json:"evidence_artifact"`
	RiidoTask            string           `json:"riido_task"`
	StoreContract        string           `json:"store_contract"`
	CompatibilityMarkers []string         `json:"compatibility_markers"`
	Decisions            []string         `json:"decisions"`
	ExternalSources      []externalSource `json:"external_sources"`
	DaemonChanges        []workRow        `json:"daemon_changes"`
	ServerChanges        []workRow        `json:"server_changes"`
	ReviewNotes          []string         `json:"review_notes"`
}
