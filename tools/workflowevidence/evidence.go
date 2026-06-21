package main

func newEvidence(m manifest, result auditResult) evidence {
	status := "verified"
	if len(result.Unregistered) > 0 || len(result.NonStrict) > 0 ||
		len(result.MissingEvidence) > 0 || len(result.MissingEvidenceTools) > 0 ||
		len(result.MissingEvidenceToolBindings) > 0 || len(result.AcceptedUnused) > 0 {
		status = "failed"
	}
	unregistered := append([]string{}, result.Unregistered...)
	nonStrict := append([]string{}, result.NonStrict...)
	missingEvidence := append([]string{}, result.MissingEvidence...)
	missingEvidenceTools := append([]string{}, result.MissingEvidenceTools...)
	missingEvidenceToolBindings := append([]string{}, result.MissingEvidenceToolBindings...)
	acceptedUnused := append([]string{}, result.AcceptedUnused...)
	return evidence{
		SchemaVersion:               evidenceSchema,
		ID:                          m.ID,
		Status:                      status,
		WorkflowCount:               len(result.Records),
		CoveredCount:                result.Covered,
		AcceptedGapCount:            result.Accepted,
		EvidenceToolCount:           result.EvidenceTools,
		EvidenceToolCoveredCount:    result.EvidenceToolCovered,
		EvidenceToolBoundCount:      result.EvidenceToolBound,
		MissingEvidenceTools:        missingEvidenceTools,
		MissingEvidenceToolBindings: missingEvidenceToolBindings,
		NonStrict:                   nonStrict,
		MissingEvidence:             missingEvidence,
		Unregistered:                unregistered,
		AcceptedUnused:              acceptedUnused,
		Records:                     result.Records,
		Loop:                        m.Loop,
	}
}
