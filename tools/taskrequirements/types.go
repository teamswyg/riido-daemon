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
	Loop             evidenceLoop  `json:"loop"`
	Purpose          string        `json:"purpose"`
	Surfaces         []Surface     `json:"surfaces"`
	Inputs           []Input       `json:"inputs"`
	SourceChecks     []SourceCheck `json:"source_checks"`
	Assertions       []string      `json:"assertions"`
}

type Surface struct {
	Name               string   `json:"name"`
	Meaning            string   `json:"meaning"`
	CapabilityFlag     string   `json:"capability_flag"`
	SchedulingConstant string   `json:"scheduling_constant"`
	CandidateField     string   `json:"candidate_field"`
	SourceChecks       []string `json:"source_checks"`
}

type Input struct {
	Name         string   `json:"name"`
	Summary      string   `json:"summary"`
	SourceChecks []string `json:"source_checks"`
}

type SourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}
