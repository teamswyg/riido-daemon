package bridge

import (
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

// TaskRequest is the provider-neutral input to Run.
type TaskRequest struct {
	ID           string
	Provider     Provider
	Prompt       string
	Cwd          string
	Model        string
	SystemPrompt string
	MaxTurns     int

	// RequiredSurfaces names provider-neutral surfaces that must be
	// present before the daemon scheduler may execute the task.
	RequiredSurfaces []string `json:"required_surfaces,omitempty"`
	// AllowExperimentalRuntime opts this task into runtimes whose
	// capability snapshot requires explicit experimental use.
	AllowExperimentalRuntime bool `json:"allow_experimental_runtime,omitempty"`

	Timeout         time.Duration
	SemanticIdle    time.Duration
	ResumeSessionID string
	Worktree        *assignmentcontract.AssignmentWorktree
	Env             map[string]string
	CustomArgs      []string
	MCPConfig       []byte
	Metadata        map[string]string
}
