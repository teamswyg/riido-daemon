package main

type manifest struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	GeneratedDocs    generatedDocs `json:"generated_docs"`
	EvidenceArtifact string        `json:"evidence_artifact"`
	Intro            string        `json:"intro"`
	ReadOrder        []readEntry   `json:"read_order"`
	Decisions        []decision    `json:"decisions"`
	Repos            []repo        `json:"repos"`
	Rules            []string      `json:"rules"`
}

type generatedDocs struct {
	Readme      string `json:"readme"`
	DocumentMap string `json:"document_map"`
}

type readEntry struct {
	Doc         string `json:"doc"`
	Description string `json:"description"`
}

type decision struct {
	Topic string   `json:"topic"`
	Docs  []string `json:"docs"`
}

type repo struct {
	Repo           string `json:"repo"`
	Responsibility string `json:"responsibility"`
}

type evidenceFile struct {
	SchemaVersion  string   `json:"schema_version"`
	ID             string   `json:"id"`
	ObservedAt     string   `json:"observed_at"`
	Status         string   `json:"status"`
	GeneratedDocs  []string `json:"generated_docs"`
	ReadOrderCount int      `json:"read_order_count"`
	DecisionCount  int      `json:"decision_count"`
	RepoCount      int      `json:"repo_count"`
	RuleCount      int      `json:"rule_count"`
	Assertions     []string `json:"assertions"`
}
