package main

type manifest struct {
	SchemaVersion      string             `json:"schema_version"`
	ID                 string             `json:"id"`
	Title              string             `json:"title"`
	GeneratedDoc       string             `json:"generated_doc"`
	Workflow           string             `json:"workflow"`
	EvidenceArtifact   string             `json:"evidence_artifact"`
	ScanRoots          []string           `json:"scan_roots"`
	ScanFiles          []string           `json:"scan_files"`
	ManualGroups       []manualGroup      `json:"manual_groups"`
	ManifestLoopBudget manifestLoopBudget `json:"manifest_loop_budget"`
	Assertions         []string           `json:"assertions"`
	Loop               evidenceLoop       `json:"loop"`
}

type manualGroup struct {
	ID           string   `json:"id"`
	Owner        string   `json:"owner"`
	Reason       string   `json:"reason"`
	NextArtifact string   `json:"next_artifact"`
	Paths        []string `json:"paths,omitempty"`
	PathPrefixes []string `json:"path_prefixes,omitempty"`
}

type docClass struct {
	Path      string
	Kind      string
	Group     string
	Reason    string
	Generator string
}

type manifestLoopBudget struct {
	MaxMissing        int            `json:"max_missing"`
	MaxMissingByGroup map[string]int `json:"max_missing_by_group"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

type manualDir struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

type manualSample struct {
	Group  string `json:"group"`
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

type generatedOrigin struct {
	Generator string   `json:"generator"`
	Count     int      `json:"count"`
	Samples   []string `json:"samples"`
}

type generatedOriginWorkflowCoverage struct {
	CoveredCount int                           `json:"covered_count"`
	MissingCount int                           `json:"missing_count"`
	Missing      []generatedOriginWorkflowMiss `json:"missing"`
}

type generatedOriginWorkflowMiss struct {
	Generator string `json:"generator"`
	Tool      string `json:"tool"`
	Count     int    `json:"count"`
}
