package bridge

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
	"github.com/teamswyg/riido-daemon/internal/execution"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

// TaskRequest is the provider-neutral input to Run.
type TaskRequest struct {
	ReferenceFolders *assignmentcontract.ReferenceFolders
	// LocalExecutionScope is source-created, never decoded from task/provider input.
	LocalExecutionScope *executionproof.Scope `json:"-"`

	ExpectedCapabilityFingerprint string `json:"-"`
	RequireHealthyCapability      bool   `json:"-"`

	RecoveryPolicy *assignmentcontract.ProviderRecoveryPolicy `json:"recovery_policy,omitempty"`
	// Original neutral input is retained only in memory for an explicitly authorized handoff.
	ProviderNeutralPrompt string `json:"-"`
	AgentInstruction      string `json:"-"`

	// Host bindings are captured once during preparation and never decoded from task input.
	HostExecutionBinding         *execution.Binding `json:"-"`
	HostExecutionBindingSelected bool               `json:"-"`

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
