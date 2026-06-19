package main

type options struct {
	Repo        string
	Manifest    string
	Doc         string
	Write       bool
	Check       bool
	EvidenceOut string
}

type manifest struct {
	SchemaVersion  string   `json:"schema_version"`
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	GeneratedDoc   string   `json:"generated_doc"`
	RequiredPhases []string `json:"required_phases"`
	LoopFiles      []string `json:"loop_files"`
	Loops          []loop   `json:"loops"`
	OpenGaps       []gap    `json:"open_gaps"`
}

type loop struct {
	ID            string     `json:"id"`
	Owner         string     `json:"owner"`
	Observation   phase      `json:"observation"`
	Hypothesis    phase      `json:"hypothesis"`
	Execution     phase      `json:"execution"`
	Evaluation    phase      `json:"evaluation"`
	Retrospective phase      `json:"retrospective"`
	Evidence      []evidence `json:"evidence"`
}

type phase struct {
	Summary   string   `json:"summary"`
	Artifacts []string `json:"artifacts"`
}

type evidence struct {
	Kind   string `json:"kind"`
	Ref    string `json:"ref"`
	Proves string `json:"proves"`
}

type gap struct {
	ID                   string `json:"id"`
	Owner                string `json:"owner"`
	CurrentHandling      string `json:"current_handling"`
	RequiredNextArtifact string `json:"required_next_artifact"`
}

type evidenceReport struct {
	SchemaVersion     string   `json:"schema_version"`
	ID                string   `json:"id"`
	Status            string   `json:"status"`
	GeneratedDoc      string   `json:"generated_doc"`
	LoopCount         int      `json:"loop_count"`
	OpenGapCount      int      `json:"open_gap_count"`
	EvidenceItemCount int      `json:"evidence_item_count"`
	RequiredPhases    []string `json:"required_phases"`
	ProblemSummaries  []string `json:"problem_summaries,omitempty"`
}
