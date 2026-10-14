package main

type qaScheduleEvidence struct {
	ID                      string `json:"id"`
	Cadence                 string `json:"cadence"`
	FreshnessWindow         string `json:"freshness_window"`
	Entrypoint              string `json:"entrypoint"`
	EvidenceCount           int    `json:"evidence_count"`
	ClosedLoopStepCount     int    `json:"closed_loop_step_count"`
	RejectConditionCount    int    `json:"reject_condition_count"`
	LocalQARunEvidenceBound bool   `json:"local_qa_run_evidence_bound"`
	ScheduleEvidenceBound   bool   `json:"schedule_evidence_bound"`
	DashboardHandoffBound   bool   `json:"dashboard_handoff_bound"`
	Status                  string `json:"status"`
	PartialReason           string `json:"partial_reason,omitempty"`
}
