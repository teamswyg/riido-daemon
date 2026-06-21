package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type manifest struct {
	SchemaVersion    string         `json:"schema_version"`
	LoopSource       string         `json:"loop_source,omitempty"`
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	GeneratedDoc     string         `json:"generated_doc"`
	Workflow         string         `json:"workflow"`
	EvidenceArtifact string         `json:"evidence_artifact"`
	RiidoTask        string         `json:"riido_task"`
	ModulePath       string         `json:"module_path"`
	BinaryPackage    string         `json:"binary_package"`
	Purpose          string         `json:"purpose"`
	Decisions        []string       `json:"decisions"`
	PackageRolesFile string         `json:"package_roles_file"`
	ImportRulesFile  string         `json:"import_rules_file"`
	PortsFile        string         `json:"ports_file"`
	PackageRoles     []packageRole  `json:"package_roles"`
	ImportRules      []importRule   `json:"import_rules"`
	Ports            []port         `json:"ports"`
	FactorBoundary   factorBoundary `json:"factor_boundary"`
	ChangeProcedure  []string       `json:"change_procedure"`
	DetailDocs       []detailDoc    `json:"detail_docs"`
	Assertions       []string       `json:"assertions"`
}
