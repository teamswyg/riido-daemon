package execution

// Observation is host-only. It is never assignment metadata or a SaaS DTO.
type Observation struct {
	AssignmentID    string `json:"assignment_id"`
	WorkspaceID     string `json:"workspace_id,omitempty"`
	TaskID          string `json:"task_id,omitempty"`
	RuntimeID       string `json:"runtime_id"`
	RunID           string `json:"run_id,omitempty"`
	Directory       string `json:"directory,omitempty"`
	SourceDirectory string `json:"source_directory,omitempty"`
	Phase           string `json:"phase"`
}
