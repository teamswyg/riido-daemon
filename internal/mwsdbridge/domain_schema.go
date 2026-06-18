package mwsdbridge

type DomainExport struct {
	SchemaVersion string             `json:"schema_version"`
	Path          string             `json:"path"`
	Domain        string             `json:"domain"`
	Repositories  []DomainRepository `json:"repositories"`
	Diagnostics   []Diagnostic       `json:"diagnostics"`
}

type DomainRepository struct {
	Name       string   `json:"name"`
	Owner      string   `json:"owner"`
	Visibility string   `json:"visibility"`
	SSOTScope  string   `json:"ssot_scope"`
	LocalPath  string   `json:"local_path"`
	Remote     string   `json:"remote"`
	Role       string   `json:"role"`
	Consumes   []string `json:"consumes"`
}
