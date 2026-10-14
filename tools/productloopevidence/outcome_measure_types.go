package main

type outcomeMeasure struct {
	ID                                string                    `json:"id"`
	ScenarioIDs                       []string                  `json:"scenario_ids"`
	Linked                            bool                      `json:"linked"`
	OutcomeEvidenceLinked             bool                      `json:"outcome_evidence_linked"`
	MissingScenarioIDs                []string                  `json:"missing_scenario_ids,omitempty"`
	MissingOutcomeEvidenceScenarioIDs []string                  `json:"missing_outcome_evidence_scenario_ids,omitempty"`
	ScenarioEvidence                  []outcomeScenarioEvidence `json:"scenario_evidence"`
}

type outcomeScenarioEvidence struct {
	ID                     string `json:"id"`
	LocalAcceptancePresent bool   `json:"local_acceptance_present"`
	RunStatus              string `json:"run_status,omitempty"`
	OutcomeEvidenceLinked  bool   `json:"outcome_evidence_linked"`
}
