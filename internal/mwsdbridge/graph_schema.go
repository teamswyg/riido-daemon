package mwsdbridge

type GraphExport struct {
	SchemaVersion string     `json:"schema_version"`
	Root          string     `json:"root"`
	Documents     []Document `json:"documents"`
	Stats         GraphStats `json:"stats"`
}

type Document struct {
	Path                string   `json:"path"`
	ID                  string   `json:"id"`
	Title               string   `json:"title"`
	Status              string   `json:"status"`
	Owner               string   `json:"owner"`
	Links               []string `json:"links"`
	Backlinks           []string `json:"backlinks"`
	MissingLinks        []string `json:"missing_links"`
	HasBacklinksSection bool     `json:"has_backlinks_section"`
}

type GraphStats struct {
	DocumentCount       int `json:"document_count"`
	NodeCount           int `json:"node_count"`
	EdgeCount           int `json:"edge_count"`
	DiagnosticCount     int `json:"diagnostic_count"`
	ErrorCount          int `json:"error_count"`
	WarningCount        int `json:"warning_count"`
	UnresolvedLinkCount int `json:"unresolved_link_count"`
}
