package main

type figmaSection struct {
	Refs        []string `json:"refs"`
	DaemonScope string   `json:"daemon_scope"`
	NotOwned    []string `json:"not_owned"`
}

type splitRepoFragment struct {
	SchemaVersion           string   `json:"schema_version"`
	LoopSource              string   `json:"loop_source"`
	Rules                   []string `json:"rules"`
	DaemonMustNotRedefine   []string `json:"daemon_must_not_redefine"`
	DownstreamBoundaryCheck string   `json:"downstream_boundary_check"`
	DownstreamBoundaryNote  string   `json:"downstream_boundary_note"`
}

type changeFragment struct {
	SchemaVersion string   `json:"schema_version"`
	LoopSource    string   `json:"loop_source"`
	Summary       string   `json:"summary"`
	SamePRUpdates []string `json:"same_pr_updates"`
}

type evidence struct {
	SchemaVersion      string   `json:"schema_version"`
	ID                 string   `json:"id"`
	Status             string   `json:"status"`
	GeneratedDocs      []string `json:"generated_docs"`
	ContextCount       int      `json:"context_count"`
	ACLCount           int      `json:"acl_count"`
	FigmaBoundaryCount int      `json:"figma_boundary_count"`
	ProblemSummaries   []string `json:"problem_summaries,omitempty"`
	EvidenceArtifact   string   `json:"evidence_artifact"`
}
