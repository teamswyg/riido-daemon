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
	Inputs           []Input       `json:"inputs"`
	Gates            []Gate        `json:"gates"`
	Passthroughs     []Input       `json:"passthroughs"`
	AbsentScans      []AbsentScan  `json:"absent_scans"`
	SourceChecks     []SourceCheck `json:"source_checks"`
	Assertions       []string      `json:"assertions"`
}

type Input struct {
	Name         string   `json:"name"`
	Source       string   `json:"source,omitempty"`
	Summary      string   `json:"summary,omitempty"`
	SourceChecks []string `json:"source_checks"`
}

type Gate struct {
	Order        int      `json:"order"`
	Code         string   `json:"code"`
	Summary      string   `json:"summary"`
	SourceChecks []string `json:"source_checks"`
}

type SourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}

type AbsentScan struct {
	Name   string   `json:"name"`
	Scope  []string `json:"scope"`
	Tokens []string `json:"tokens"`
	Reason string   `json:"reason"`
}
