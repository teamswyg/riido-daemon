package main

type Evidence struct {
	ID             string           `json:"id"`
	Artifact       string           `json:"artifact"`
	Problems       []string         `json:"problems"`
	SourceChecks   []SourceEvidence `json:"source_checks"`
	AbsentSurfaces []AbsentEvidence `json:"absent_surfaces"`
	Assertions     []string         `json:"assertions"`
}

type SourceEvidence struct {
	Name string `json:"name"`
	File string `json:"file"`
	OK   bool   `json:"ok"`
}

type AbsentEvidence struct {
	Name  string `json:"name"`
	Scope string `json:"scope"`
	Token string `json:"token"`
	OK    bool   `json:"ok"`
}
