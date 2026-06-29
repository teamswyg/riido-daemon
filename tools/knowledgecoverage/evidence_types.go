package main

type evidence struct {
	SchemaVersion              string                          `json:"schema_version"`
	ID                         string                          `json:"id"`
	Status                     string                          `json:"status"`
	ScannedCount               int                             `json:"scanned_count"`
	GeneratedCount             int                             `json:"generated_count"`
	DirectSSOTCount            int                             `json:"direct_ssot_count"`
	ManualCount                int                             `json:"manual_count"`
	GeneratedOriginCount       int                             `json:"generated_origin_count"`
	GeneratedWorkflowCovered   int                             `json:"generated_workflow_covered_count"`
	GeneratedWorkflowMissing   int                             `json:"generated_workflow_missing_count"`
	ManualGroups               []string                        `json:"manual_groups"`
	ManualByGroup              map[string]int                  `json:"manual_by_group"`
	ManualTopDirs              []manualDir                     `json:"manual_top_dirs"`
	ManualSamples              []manualSample                  `json:"manual_samples"`
	GeneratedOrigins           []generatedOrigin               `json:"generated_origins"`
	GeneratedWorkflowCoverage  generatedOriginWorkflowCoverage `json:"generated_workflow_coverage"`
	ManifestInventory          manifestInventory               `json:"manifest_inventory"`
	ManifestLoopCount          int                             `json:"manifest_loop_count"`
	ManifestDirectLoopCount    int                             `json:"manifest_direct_loop_count"`
	ManifestDelegatedLoopCount int                             `json:"manifest_delegated_loop_count"`
	ManifestMissingLoopCount   int                             `json:"manifest_missing_loop_count"`
	ManifestMissingLoopGroups  []manifestGroupCount            `json:"manifest_missing_loop_groups"`
	ManifestMissingLoopSamples []manifestGroupSample           `json:"manifest_missing_loop_samples"`
	ManifestLoopBudget         manifestLoopBudget              `json:"manifest_loop_budget"`
	LoopRegistry               loopRegistrySummary             `json:"loop_registry"`
	ProblemCount               int                             `json:"problem_count"`
	ProblemSummaries           []string                        `json:"problem_summaries"`
	EvidenceArtifact           string                          `json:"evidence_artifact"`
	Loop                       evidenceLoop                    `json:"loop"`
}
