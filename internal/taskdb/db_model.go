package taskdb

// TaskDB is Riido's first local task transition database.
//
// It stays dependency-light on purpose: the local daemon source is single host
// and local-only, so atomic JSON replacement gives us a simple, inspectable
// database before we introduce a heavier embedded store.
type TaskDB struct {
	SchemaVersion          string                     `json:"schema_version"`
	ProjectionVersion      string                     `json:"projection_version"`
	Root                   string                     `json:"root"`
	Domain                 string                     `json:"domain"`
	UpdatedAt              string                     `json:"updated_at"`
	RecommendedProvider    string                     `json:"recommended_provider"`
	RecommendedDecisionLLM string                     `json:"recommended_decision_llm"`
	DecisionGate           string                     `json:"decision_gate"`
	ProviderCandidates     []ProviderCandidate        `json:"provider_candidates"`
	Tasks                  []TaskRecord               `json:"tasks"`
	Transitions            []TaskTransitionRecord     `json:"transitions"`
	Evidence               []TaskEvidenceRecord       `json:"evidence"`
	CommandReceipts        []TaskCommandReceiptRecord `json:"command_receipts"`
	Diagnostics            []ProjectionDiagnostic     `json:"diagnostics"`
}

func EmptyTaskDB() TaskDB {
	return TaskDB{
		SchemaVersion:   TaskDBSchemaVersion,
		Tasks:           []TaskRecord{},
		Transitions:     []TaskTransitionRecord{},
		Evidence:        []TaskEvidenceRecord{},
		CommandReceipts: []TaskCommandReceiptRecord{},
		Diagnostics:     []ProjectionDiagnostic{},
	}
}
