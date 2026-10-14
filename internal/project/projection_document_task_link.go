package project

// DocumentTaskLink is the first deterministic bridge from document SSOT to
// Riido task identity. It is not yet a persisted task row; it is the stable
// source mapping the future task store can consume.
type DocumentTaskLink struct {
	TaskID                 string `json:"task_id"`
	DocumentID             string `json:"document_id"`
	DocumentPath           string `json:"document_path"`
	Title                  string `json:"title"`
	Status                 string `json:"status"`
	Owner                  string `json:"owner"`
	ProjectID              string `json:"project_id"`
	RecommendedProvider    string `json:"recommended_provider"`
	RecommendedDecisionLLM string `json:"recommended_decision_llm"`
	RequiresHumanApproval  bool   `json:"requires_human_approval"`
	HarnessNextDirection   string `json:"harness_next_direction"`
}
