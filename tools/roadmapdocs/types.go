package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type manifest struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	GeneratedDoc     string        `json:"generated_doc"`
	Workflow         string        `json:"workflow"`
	EvidenceArtifact string        `json:"evidence_artifact"`
	RiidoTask        string        `json:"riido_task"`
	Intro            []string      `json:"intro"`
	Questions        []question    `json:"questions"`
	PromotionRule    string        `json:"promotion_rule"`
	SourceChecks     []sourceCheck `json:"source_checks"`
	Assertions       []string      `json:"assertions"`
}

type question struct {
	ID              string `json:"id"`
	Area            string `json:"area"`
	Question        string `json:"question"`
	CurrentHandling string `json:"current_handling"`
}

type sourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}

type evidence struct {
	SchemaVersion    string              `json:"schema_version"`
	ID               string              `json:"id"`
	Status           string              `json:"status"`
	GeneratedDoc     string              `json:"generated_doc"`
	QuestionCount    int                 `json:"question_count"`
	SourceChecks     []sourceCheckResult `json:"source_checks"`
	AssertionCount   int                 `json:"assertion_count"`
	ProblemSummaries []string            `json:"problem_summaries,omitempty"`
	EvidenceArtifact string              `json:"evidence_artifact"`
}

type sourceCheckResult struct {
	Name   string `json:"name"`
	File   string `json:"file"`
	Passed bool   `json:"passed"`
}
