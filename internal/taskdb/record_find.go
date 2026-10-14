package taskdb

func findTransitionByID(transitions []TaskTransitionRecord, id string) (TaskTransitionRecord, bool) {
	for _, transition := range transitions {
		if transition.ID == id {
			return transition, true
		}
	}
	return TaskTransitionRecord{}, false
}

func findEvidenceByID(evidenceRecords []TaskEvidenceRecord, id string) (TaskEvidenceRecord, bool) {
	for _, evidence := range evidenceRecords {
		if evidence.ID == id {
			return evidence, true
		}
	}
	return TaskEvidenceRecord{}, false
}
