package main

type Evidence struct {
	ID             string             `json:"id"`
	Artifact       string             `json:"artifact"`
	Problems       []string           `json:"problems"`
	SourceChecks   []SourceEvidence   `json:"source_checks"`
	PolicyChecks   []PolicyEvidence   `json:"policy_checks"`
	CodexArgChecks []CodexArgEvidence `json:"codex_arg_checks"`
	Assertions     []string           `json:"assertions"`
}

type SourceEvidence struct {
	Name string `json:"name"`
	File string `json:"file"`
	OK   bool   `json:"ok"`
}

type PolicyEvidence struct {
	Surface string `json:"surface"`
	OK      bool   `json:"ok"`
}

type CodexArgEvidence struct {
	Arg string `json:"arg"`
	OK  bool   `json:"ok"`
}
