package main

type options struct {
	Manifest    string
	EvidenceOut string
	WriteDoc    bool
	CheckDoc    bool
	Strict      bool
}

type manifest struct {
	SchemaVersion           string          `json:"schema_version"`
	ID                      string          `json:"id"`
	Title                   string          `json:"title"`
	GeneratedDoc            string          `json:"generated_doc"`
	Workflow                string          `json:"workflow"`
	EvidenceArtifact        string          `json:"evidence_artifact"`
	LoopRegistry            string          `json:"loop_registry"`
	EntrypointRouteMap      string          `json:"entrypoint_route_map"`
	LocalAcceptanceManifest string          `json:"local_acceptance_manifest"`
	QASystemManifest        string          `json:"qa_system_manifest"`
	LocalQARunEvidence      string          `json:"local_qa_run_evidence"`
	Thresholds              thresholds      `json:"thresholds"`
	OutcomeSignals          []outcomeSignal `json:"outcome_signals"`
	Loop                    evidenceLoop    `json:"loop"`
}

type thresholds struct {
	MaxEntrypointsBeforePartial int `json:"max_entrypoints_before_partial"`
	StalePartialAfterDays       int `json:"stale_partial_after_days"`
}

type outcomeSignal struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	ScenarioIDs []string `json:"scenario_ids"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
