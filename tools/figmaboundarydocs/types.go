package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type manifest struct {
	SchemaVersion                     string              `json:"schema_version"`
	ID                                string              `json:"id"`
	RiidoTask                         string              `json:"riido_task"`
	HardeningTasks                    []string            `json:"hardening_tasks"`
	HumanDoc                          string              `json:"human_doc"`
	SourceCoverageManifest            string              `json:"source_coverage_manifest"`
	SourceCoverageManifestProvenance  upstreamManifestRef `json:"source_coverage_manifest_provenance"`
	MirroredSupportingToolLimitations []toolLimitation    `json:"mirrored_supporting_tool_limitations"`
	Figma                             figmaRef            `json:"figma"`
	BoundaryPolicy                    boundaryPolicy      `json:"boundary_policy"`
	EntryFiles                        []string            `json:"entry_files"`
	Entries                           []boundaryEntry     `json:"-"`
}

type upstreamManifestRef struct {
	Repo                    string   `json:"repo"`
	Path                    string   `json:"path"`
	SchemaVersion           string   `json:"schema_version"`
	ID                      string   `json:"id"`
	MirrorsSourceField      string   `json:"mirrors_source_field"`
	SourceFieldIntroducedBy string   `json:"source_field_introduced_by"`
	StabilizedBy            []string `json:"stabilized_by"`
}
