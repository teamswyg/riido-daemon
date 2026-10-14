package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type manifest struct {
	SchemaVersion    string       `json:"schema_version"`
	ID               string       `json:"id"`
	Title            string       `json:"title"`
	LoopSource       string       `json:"loop_source,omitempty"`
	GeneratedDoc     string       `json:"generated_doc"`
	Workflow         string       `json:"workflow"`
	EvidenceArtifact string       `json:"evidence_artifact"`
	RiidoTask        string       `json:"riido_task"`
	AssignmentFSMDoc string       `json:"assignment_fsm_doc"`
	EvidenceManifest string       `json:"evidence_manifest"`
	Fragments        fragmentRefs `json:"fragments"`
}

type fragmentRefs struct {
	Overview       string `json:"overview"`
	RiskModel      string `json:"risk_model"`
	ExecutionModel string `json:"execution_model"`
	LifecycleModel string `json:"lifecycle_model"`
	Governance     string `json:"governance"`
}
