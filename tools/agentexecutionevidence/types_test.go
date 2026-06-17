package agentexecutionevidence

type evidenceManifest struct {
	SchemaVersion       string              `json:"schema_version"`
	ID                  string              `json:"id"`
	RiidoTask           string              `json:"riido_task"`
	HumanDoc            string              `json:"human_doc"`
	SourceDocuments     []string            `json:"source_documents"`
	LocalEvidence       []localEvidence     `json:"local_evidence"`
	ExternalEvidence    []externalEvidence  `json:"external_evidence"`
	RemainingBoundaries []remainingBoundary `json:"remaining_boundaries"`
}

type localEvidence struct {
	Risk    string `json:"risk"`
	Status  string `json:"status"`
	Package string `json:"package"`
	Test    string `json:"test"`
	Proves  string `json:"proves"`
}

type externalEvidence struct {
	Risk   string `json:"risk"`
	Status string `json:"status"`
	Repo   string `json:"repo"`
	Test   string `json:"test"`
	Proves string `json:"proves"`
}

type remainingBoundary struct {
	ID                   string `json:"id"`
	Owner                string `json:"owner"`
	CurrentHandling      string `json:"current_handling"`
	RequiredNextArtifact string `json:"required_next_artifact"`
}
