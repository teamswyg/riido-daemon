package taskdb

func replayExistingTaskEvidence(db TaskDB, input TaskEvidenceInput, actor, source string) (TaskEvidenceRecord, TaskCommandReceiptRecord, bool, error) {
	receipt, found, err := findCommandReceiptByCommandID(db, input.Guard.CommandID)
	if err != nil || !found {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, found, err
	}
	if err := validateCommandReceiptReplay(receipt, "evidence", input.TaskID, actor, source, input.Guard); err != nil {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, err
	}
	evidence, err := findReplayedEvidence(db, receipt)
	if err != nil {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, err
	}
	if err := validateReplayedEvidence(receipt, evidence, input, actor, source); err != nil {
		return TaskEvidenceRecord{}, TaskCommandReceiptRecord{}, true, err
	}
	return evidence, receipt, true, nil
}
