package main

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
