package taskdb

func validateReplayedEvidenceActorFields(commandID string, evidence TaskEvidenceRecord, input TaskEvidenceInput, actor, source string) error {
	if evidence.Actor != actor {
		return commandReplayMismatch(commandID, "actor")
	}
	if evidence.Source != source {
		return commandReplayMismatch(commandID, "source")
	}
	if evidence.Summary != input.Summary {
		return commandReplayMismatch(commandID, "summary")
	}
	return nil
}
