package figmaboundary

type boundaryManifest struct {
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
	Entries                           []boundaryEntry     `json:"entries"`
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

type toolLimitation struct {
	SourceID                   string   `json:"source_id"`
	SourceOwner                string   `json:"source_owner"`
	SourceStabilizedBy         []string `json:"source_stabilized_by"`
	LocalRiidoTask             string   `json:"local_riido_task"`
	DaemonScope                string   `json:"daemon_scope"`
	RequiredAuthoritativePages []string `json:"required_authoritative_pages"`
	MustPreserveNonUINodes     []string `json:"must_preserve_non_ui_nodes"`
}

type figmaRef struct {
	FileKey     string `json:"file_key"`
	FileName    string `json:"file_name"`
	PageID      string `json:"page_id"`
	PageName    string `json:"page_name"`
	InspectedAt string `json:"inspected_at"`
}

type boundaryPolicy struct {
	Summary  string `json:"summary"`
	TopDown  string `json:"top_down"`
	BottomUp string `json:"bottom_up"`
}

type boundaryEntry struct {
	NodeID              string   `json:"node_id"`
	Name                string   `json:"name"`
	UpstreamOwner       []string `json:"upstream_owner"`
	DaemonScope         string   `json:"daemon_scope"`
	DaemonConsumedFacts []string `json:"daemon_consumed_facts"`
	ClientOwnedFacts    []string `json:"client_owned_facts"`
}
