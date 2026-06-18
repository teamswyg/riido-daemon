package taskdb

func normalizeTaskDB(db TaskDB) TaskDB {
	defaultTaskDBCollections(&db)
	sortTaskDB(&db)
	recountTaskDB(&db)
	return db
}

func defaultTaskDBCollections(db *TaskDB) {
	if db.SchemaVersion == "" {
		db.SchemaVersion = TaskDBSchemaVersion
	}
	if db.Tasks == nil {
		db.Tasks = []TaskRecord{}
	}
	if db.Transitions == nil {
		db.Transitions = []TaskTransitionRecord{}
	}
	if db.Evidence == nil {
		db.Evidence = []TaskEvidenceRecord{}
	}
	if db.CommandReceipts == nil {
		db.CommandReceipts = []TaskCommandReceiptRecord{}
	}
	if db.Diagnostics == nil {
		db.Diagnostics = []ProjectionDiagnostic{}
	}
	if db.ProviderCandidates == nil {
		db.ProviderCandidates = []ProviderCandidate{}
	}
}
