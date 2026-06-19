package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type manifest struct {
	SchemaVersion         string        `json:"schema_version"`
	ID                    string        `json:"id"`
	Title                 string        `json:"title"`
	GeneratedDoc          string        `json:"generated_doc"`
	Workflow              string        `json:"workflow"`
	EvidenceArtifact      string        `json:"evidence_artifact"`
	Purpose               string        `json:"purpose"`
	ReleaseWorkflow       string        `json:"release_workflow"`
	BuildScript           string        `json:"build_script"`
	PublishScript         string        `json:"publish_script"`
	InstallScript         string        `json:"install_script"`
	Targets               []target      `json:"targets"`
	ArchiveContents       []string      `json:"archive_contents"`
	Installer             installer     `json:"installer"`
	DesktopMSIX           desktopMSIX   `json:"desktop_msix"`
	ForbiddenArchiveItems []string      `json:"forbidden_archive_items"`
	InheritedRefs         []string      `json:"inherited_refs"`
	SourceChecks          []sourceCheck `json:"source_checks"`
	DetailDocs            []detailDoc   `json:"detail_docs"`
	Assertions            []string      `json:"assertions"`
}

type target struct {
	Platform string `json:"platform"`
	GOOS     string `json:"goos"`
	GOARCH   string `json:"goarch"`
	Format   string `json:"format"`
}

type detailDoc struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}
