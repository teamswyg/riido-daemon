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
	LoopSource       string        `json:"loop_source,omitempty"`
	GeneratedDoc     string        `json:"generated_doc"`
	Workflow         string        `json:"workflow"`
	EvidenceArtifact string        `json:"evidence_artifact"`
	Purpose          string        `json:"purpose"`
	SearchOrder      []SearchStep  `json:"search_order"`
	Rules            []Rule        `json:"rules"`
	SourceChecks     []SourceCheck `json:"source_checks"`
	Assertions       []string      `json:"assertions"`
}

type SearchStep struct {
	Order        int      `json:"order"`
	Name         string   `json:"name"`
	Summary      string   `json:"summary"`
	SourceChecks []string `json:"source_checks"`
	Behavior     string   `json:"behavior,omitempty"`
}

type Rule struct {
	Name         string   `json:"name"`
	Summary      string   `json:"summary"`
	SourceChecks []string `json:"source_checks"`
	Behavior     string   `json:"behavior,omitempty"`
}

type SourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}
