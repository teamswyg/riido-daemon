package main

type report struct {
	SchemaVersion     string               `json:"schema_version"`
	ID                string               `json:"id"`
	Status            string               `json:"status"`
	GeneratedDoc      string               `json:"generated_doc"`
	Workflow          string               `json:"workflow"`
	EvidenceArtifact  string               `json:"evidence_artifact"`
	MetaComplexity    metaComplexity       `json:"meta_complexity"`
	ProductAcceptance productAcceptance    `json:"product_acceptance"`
	QASchedule        qaScheduleEvidence   `json:"qa_schedule"`
	PartialReduction  partialReduction     `json:"partial_reduction"`
	Candidates        []closedLoopCandiate `json:"closed_loop_candidates"`
	ProblemCount      int                  `json:"problem_count"`
	Problems          []string             `json:"problems"`
}

type metaComplexity struct {
	ToolEntrypointCount int             `json:"tool_entrypoint_count"`
	WorkflowCount       int             `json:"workflow_count"`
	VerifierFileCount   int             `json:"verifier_file_count"`
	GeneratedDocCount   int             `json:"generated_doc_count"`
	EntrypointCount     int             `json:"entrypoint_count"`
	EntrypointThreshold int             `json:"entrypoint_threshold"`
	MappingCoverage     mappingCoverage `json:"mapping_coverage"`
	RouteCoverage       routeCoverage   `json:"route_coverage"`
	Status              string          `json:"status"`
	PartialReason       string          `json:"partial_reason,omitempty"`
}

type mappingCoverage struct {
	ClaimCount             int     `json:"claim_count"`
	ClaimWithVerifierCount int     `json:"claim_with_verifier_count"`
	BoundFileCount         int     `json:"bound_file_count"`
	DeclaredVerifierCount  int     `json:"declared_verifier_count"`
	CoverageRatio          float64 `json:"coverage_ratio"`
}

type routeCoverage struct {
	RouteCount               int                `json:"route_count"`
	EntrypointCount          int                `json:"entrypoint_count"`
	CoveredEntrypointCount   int                `json:"covered_entrypoint_count"`
	UncoveredEntrypointCount int                `json:"uncovered_entrypoint_count"`
	CoverageRatio            float64            `json:"coverage_ratio"`
	UncoveredEntrypoints     []string           `json:"uncovered_entrypoints,omitempty"`
	Routes                   []routeCoverageRow `json:"routes"`
}

type routeCoverageRow struct {
	ID              string `json:"id"`
	Owner           string `json:"owner"`
	EntrypointCount int    `json:"entrypoint_count"`
}

type productAcceptance struct {
	ScenarioCount         int              `json:"scenario_count"`
	OutcomeSignalCount    int              `json:"outcome_signal_count"`
	LinkedSignalCount     int              `json:"linked_signal_count"`
	MissingSignalIDs      []string         `json:"missing_signal_ids"`
	MeasurementCandidates []outcomeMeasure `json:"measurement_candidates"`
	Status                string           `json:"status"`
}

type outcomeMeasure struct {
	ID          string   `json:"id"`
	ScenarioIDs []string `json:"scenario_ids"`
	Linked      bool     `json:"linked"`
}
