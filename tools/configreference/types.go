package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type Manifest struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	GeneratedDoc     string        `json:"generated_doc"`
	Workflow         string        `json:"workflow"`
	EvidenceArtifact string        `json:"evidence_artifact"`
	Purpose          string        `json:"purpose"`
	DaemonEnvSource  string        `json:"daemon_env_source"`
	Sections         []Section     `json:"sections"`
	DaemonEnvVars    []EnvVar      `json:"daemon_env_vars"`
	CoverageAnchors  []Anchor      `json:"coverage_anchors"`
	SourceChecks     []SourceCheck `json:"source_checks"`
	DetailDocs       []DetailDoc   `json:"detail_docs"`
	Assertions       []string      `json:"assertions"`
}

type Section struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

type EnvVar struct {
	Name     string `json:"name"`
	Consumer string `json:"consumer"`
	Default  string `json:"default"`
	Meaning  string `json:"meaning"`
}

type Anchor struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
	Meaning  string `json:"meaning"`
}

type SourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}
