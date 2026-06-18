package riidoapi

import "github.com/teamswyg/riido-daemon/internal/taskdb"

type TransitionRequest struct {
	TaskID      string `json:"task_id"`
	ToState     string `json:"to_state"`
	EventType   string `json:"event_type"`
	Actor       string `json:"actor"`
	Source      string `json:"source"`
	Reason      string `json:"reason"`
	Provider    string `json:"provider"`
	DecisionLLM string `json:"decision_llm"`
	ApprovalID  string `json:"approval_id"`
	CommandID   string `json:"command_id"`
}

type TransitionResponse struct {
	TaskDBPath string                          `json:"task_db_path"`
	Task       taskdb.TaskRecord               `json:"task"`
	Transition taskdb.TaskTransitionRecord     `json:"transition"`
	Receipt    taskdb.TaskCommandReceiptRecord `json:"receipt"`
}
