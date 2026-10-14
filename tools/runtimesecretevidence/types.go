package main

type manifest struct {
	SchemaVersion        string         `json:"schema_version"`
	LoopSource           string         `json:"loop_source,omitempty"`
	ID                   string         `json:"id"`
	Title                string         `json:"title"`
	GeneratedDoc         string         `json:"generated_doc"`
	Workflow             string         `json:"workflow"`
	EvidenceArtifact     string         `json:"evidence_artifact"`
	PrivateOwner         string         `json:"private_owner"`
	PublicScope          []string       `json:"public_scope"`
	EvidenceKinds        []evidenceKind `json:"evidence_kinds"`
	AllowedPacketFields  []string       `json:"allowed_packet_fields"`
	AllowedAWSOperations []string       `json:"allowed_aws_operations"`
	ForbiddenAWSOps      []string       `json:"forbidden_aws_operations"`
	ForbiddenFieldNames  []string       `json:"forbidden_field_names"`
}

type evidenceKind struct {
	ID       string   `json:"id"`
	ActualID string   `json:"actual_id"`
	Proves   []string `json:"proves"`
	Forbids  []string `json:"forbids"`
}

type evidenceFile struct {
	SchemaVersion       string   `json:"schema_version"`
	ID                  string   `json:"id"`
	ObservedAt          string   `json:"observed_at"`
	Status              string   `json:"status"`
	PrivateOwner        string   `json:"private_owner"`
	AllowedOperations   []string `json:"allowed_aws_operations"`
	ForbiddenOperations []string `json:"forbidden_aws_operations"`
	AllowedFields       []string `json:"allowed_packet_fields"`
	ForbiddenFields     []string `json:"forbidden_field_names"`
	Assertions          []string `json:"assertions"`
}
