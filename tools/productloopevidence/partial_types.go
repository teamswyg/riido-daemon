package main

type partialReduction struct {
	InferenceRequiredIDs      []string `json:"inference_required_ids"`
	ClosedLoopCandidateIDs    []string `json:"closed_loop_candidate_ids"`
	CandidateCount            int      `json:"candidate_count"`
	PromotedCount             int      `json:"promoted_count"`
	CandidateAgeUnknownCount  int      `json:"candidate_age_unknown_count"`
	StaleCandidateCount       int      `json:"stale_candidate_count"`
	StalePartialAfterDays     int      `json:"stale_partial_after_days"`
	LocalQARunEvidencePresent bool     `json:"local_qa_run_evidence_present"`
	Status                    string   `json:"status"`
}

type closedLoopCandiate struct {
	ID                    string   `json:"id"`
	Class                 string   `json:"class"`
	Reason                string   `json:"reason"`
	RequiredNextArtifacts []string `json:"required_next_artifacts"`
}
