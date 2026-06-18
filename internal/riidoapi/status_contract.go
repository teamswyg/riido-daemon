package riidoapi

type Status struct {
	SchemaVersion       string `json:"schema_version"`
	Transport           string `json:"transport"`
	SocketPath          string `json:"socket_path"`
	TaskDBPath          string `json:"task_db_path"`
	TaskDBSchemaVersion string `json:"task_db_schema_version"`
	TaskCount           int    `json:"task_count"`
	TransitionCount     int    `json:"transition_count"`
	EvidenceCount       int    `json:"evidence_count"`
	CommandReceiptCount int    `json:"command_receipt_count"`
	DiagnosticCount     int    `json:"diagnostic_count"`
	UpdatedAt           string `json:"updated_at"`
}
