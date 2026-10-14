package taskdb

func findReplayedEvidence(db TaskDB, receipt TaskCommandReceiptRecord) (TaskEvidenceRecord, error) {
	if receipt.EvidenceID == "" {
		return TaskEvidenceRecord{}, taskDBErrorf(ErrTaskDBReplay, "evidence.replay", "command_id %s replay cannot find linked evidence id", receipt.CommandID)
	}
	evidence, ok := findEvidenceByID(db.Evidence, receipt.EvidenceID)
	if !ok {
		return TaskEvidenceRecord{}, taskDBErrorf(ErrTaskDBReplay, "evidence.replay", "command_id %s replay cannot find evidence %s", receipt.CommandID, receipt.EvidenceID)
	}
	return evidence, nil
}
