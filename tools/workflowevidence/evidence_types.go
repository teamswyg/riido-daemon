package main

type evidence struct {
	SchemaVersion               string           `json:"schema_version"`
	ID                          string           `json:"id"`
	Status                      string           `json:"status"`
	WorkflowCount               int              `json:"workflow_count"`
	CoveredCount                int              `json:"covered_count"`
	StatusCounts                []statusCount    `json:"status_counts"`
	AcceptedGapCount            int              `json:"accepted_gap_count"`
	EvidenceToolCount           int              `json:"evidence_tool_count"`
	EvidenceToolCoveredCount    int              `json:"evidence_tool_covered_count"`
	EvidenceToolBoundCount      int              `json:"evidence_tool_bound_count"`
	MissingEvidenceTools        []string         `json:"missing_evidence_tools"`
	MissingEvidenceToolBindings []string         `json:"missing_evidence_tool_bindings"`
	NonStrict                   []string         `json:"non_strict_artifact_uploads"`
	MissingEvidence             []string         `json:"missing_evidence_uploads"`
	Unregistered                []string         `json:"unregistered_gaps"`
	AcceptedUnused              []string         `json:"accepted_gaps_unused"`
	Records                     []workflowRecord `json:"records"`
	Loop                        evidenceLoop     `json:"loop"`
}
