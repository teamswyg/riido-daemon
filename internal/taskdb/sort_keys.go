package taskdb

func transitionSortKey(record TaskTransitionRecord) (string, string) {
	return record.RecordedAt, record.ID
}

func evidenceSortKey(record TaskEvidenceRecord) (string, string) {
	return record.RecordedAt, record.ID
}

func receiptSortKey(record TaskCommandReceiptRecord) (string, string) {
	return record.RecordedAt, record.ID
}
