package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type manifest struct {
	SchemaVersion    string             `json:"schema_version"`
	ID               string             `json:"id"`
	Title            string             `json:"title"`
	GeneratedDoc     string             `json:"generated_doc"`
	Workflow         string             `json:"workflow"`
	EvidenceArtifact string             `json:"evidence_artifact"`
	Loop             evidenceLoop       `json:"loop"`
	RiidoTask        string             `json:"riido_task"`
	TaskTitle        string             `json:"task_title"`
	Intro            string             `json:"intro"`
	OwnershipSummary []string           `json:"ownership_summary"`
	FocusedSections  []link             `json:"focused_sections"`
	BoundaryEvidence []link             `json:"boundary_evidence"`
	RepositoryRule   string             `json:"repository_boundary_rule"`
	WorkUnitRule     string             `json:"work_unit_boundary_rule"`
	WorkUnitGate     string             `json:"work_unit_gate"`
	WorkUnitScript   string             `json:"work_unit_script"`
	Fragments        map[string]string  `json:"fragments"`
	Contexts         []contextRow       `json:"contexts"`
	ACL              aclFragment        `json:"-"`
	Dependency       dependencyFragment `json:"-"`
	FigmaDaemon      figmaFragment      `json:"-"`
	FigmaOnboarding  onboardingFragment `json:"-"`
	SplitRepo        splitRepoFragment  `json:"-"`
	ChangeProcedure  changeFragment     `json:"-"`
}

type link struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

type contextRow struct {
	ID      string `json:"id"`
	Context string `json:"context"`
	Owner   string `json:"owner"`
}
