package main

type auditResult struct {
	Records                     []workflowRecord
	Covered                     int
	Accepted                    int
	EvidenceTools               int
	EvidenceToolCovered         int
	EvidenceToolBound           int
	Unregistered                []string
	NonStrict                   []string
	MissingEvidence             []string
	MissingEvidenceTools        []string
	MissingEvidenceToolBindings []string
	AcceptedUnused              []string
}
