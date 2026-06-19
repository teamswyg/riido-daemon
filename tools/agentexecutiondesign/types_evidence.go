package main

type evidenceManifest struct {
	SchemaVersion string        `json:"schema_version"`
	ID            string        `json:"id"`
	RiidoTask     string        `json:"riido_task"`
	HumanDoc      string        `json:"human_doc"`
	SourceDocs    []string      `json:"source_documents"`
	EvidenceFiles evidenceFiles `json:"evidence_files"`
}

type evidenceFiles struct {
	Local               []string `json:"local"`
	External            []string `json:"external"`
	RemainingBoundaries []string `json:"remaining_boundaries"`
}

type evidenceItem struct {
	Risk    string `json:"risk"`
	Status  string `json:"status"`
	Repo    string `json:"repo,omitempty"`
	Package string `json:"package,omitempty"`
	Test    string `json:"test"`
	Proves  string `json:"proves"`
}

type boundaryItem struct {
	ID                   string `json:"id"`
	Owner                string `json:"owner"`
	CurrentHandling      string `json:"current_handling"`
	RequiredNextArtifact string `json:"required_next_artifact"`
}
