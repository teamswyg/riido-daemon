package taskdb

type TaskCommandReceiptRecord struct {
	ID                     string `json:"id"`
	CommandID              string `json:"command_id"`
	Kind                   string `json:"kind"`
	TaskID                 string `json:"task_id"`
	Actor                  string `json:"actor"`
	Source                 string `json:"source"`
	Provider               string `json:"provider"`
	DecisionLLM            string `json:"decision_llm"`
	ApprovalID             string `json:"approval_id"`
	DecisionGate           string `json:"decision_gate"`
	RequiresHumanApproval  bool   `json:"requires_human_approval"`
	RecommendedProvider    string `json:"recommended_provider"`
	RecommendedDecisionLLM string `json:"recommended_decision_llm"`
	HarnessNextDirection   string `json:"harness_next_direction"`
	GuardDecision          string `json:"guard_decision"`
	GuardReason            string `json:"guard_reason"`
	ReplayPolicy           string `json:"replay_policy"`
	TransitionID           string `json:"transition_id,omitempty"`
	EvidenceID             string `json:"evidence_id,omitempty"`
	Result                 string `json:"result"`
	RecordedAt             string `json:"recorded_at"`
}
