package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type Manifest struct {
	SchemaVersion         string                 `json:"schema_version"`
	ID                    string                 `json:"id"`
	Title                 string                 `json:"title"`
	GeneratedDoc          string                 `json:"generated_doc"`
	Workflow              string                 `json:"workflow"`
	EvidenceArtifact      string                 `json:"evidence_artifact"`
	Loop                  evidenceLoop           `json:"loop"`
	Purpose               string                 `json:"purpose"`
	CommandGroups         []CommandGroup         `json:"command_groups"`
	Providers             []string               `json:"providers"`
	SourceChecks          []SourceCheck          `json:"source_checks"`
	ForbiddenSourceTokens []ForbiddenSourceToken `json:"forbidden_source_tokens"`
	RelatedSections       []RelatedSection       `json:"related_sections"`
	DetailDocs            []DetailDoc            `json:"detail_docs"`
	Assertions            []string               `json:"assertions"`
}

type CommandGroup struct {
	Name         string   `json:"name"`
	Owner        string   `json:"owner"`
	Boundary     string   `json:"boundary"`
	Subcommands  []string `json:"subcommands"`
	UsageTokens  []string `json:"usage_tokens"`
	SourceChecks []string `json:"source_checks"`
}

type SourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}

type ForbiddenSourceToken struct {
	Name     string `json:"name"`
	Root     string `json:"root"`
	Suffix   string `json:"suffix"`
	Contains string `json:"contains"`
}

type RelatedSection struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}
