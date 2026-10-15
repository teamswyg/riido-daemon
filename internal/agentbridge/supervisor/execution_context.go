package supervisor

import (
	"errors"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/execution"
)

var errWorkspacePreparation = errors.New("workspace preparation failed")

func executionContextForTask(task *runningTask) assignment.ExecutionContext {
	c := assignment.ExecutionContext{WorkspaceKind: "unknown", VerificationStatus: "not_configured", DiagnosticCode: "none", NextAction: "none"}
	if task == nil {
		return c
	}
	if task.workspace != nil {
		c.WorkspaceKind = "isolated"
	}
	if task.request != nil {
		if task.request.Worktree != nil {
			c.WorkspaceKind = "assignment_worktree"
		}
		if task.request.Metadata[metadataSessionOriginRun] != "" {
			c.WorkspaceKind = "native_resume"
		}
	}
	if task.execution != nil {
		if c.WorkspaceKind != "native_resume" {
			c.WorkspaceKind = "repository_clone"
		}
		if task.execution.Binding.HasChecks() {
			c.VerificationStatus = "pending"
		}
	}
	return c
}

func resultExecutionContext(task *runningTask, res agentbridge.Result) *assignment.ExecutionContext {
	c := executionContextForTask(task)
	if res.ExecutionContext != nil {
		prior := res.ExecutionContext.Sanitized()
		c.DiagnosticCode, c.NextAction = prior.DiagnosticCode, prior.NextAction
		if prior.VerificationStatus != "unknown" {
			c.VerificationStatus = prior.VerificationStatus
		}
	}
	switch res.Status {
	case agentbridge.ResultCancelled:
		c.DiagnosticCode, c.NextAction = "run_cancelled", "none"
		if task != nil && task.verificationPending {
			c.VerificationStatus = "cancelled"
		}
	case agentbridge.ResultTimeout:
		c.DiagnosticCode, c.NextAction = "run_timeout", "review_result"
	case agentbridge.ResultAborted:
		c.DiagnosticCode, c.NextAction = "process_aborted", "review_result"
	case agentbridge.ResultBlocked, agentbridge.ResultFailed:
		if c.DiagnosticCode == "none" {
			c.DiagnosticCode, c.NextAction = "provider_failed", "check_provider"
			if res.Status == agentbridge.ResultBlocked {
				c.DiagnosticCode = "provider_blocked"
			}
		}
	case agentbridge.ResultCompleted:
		if task != nil && task.execution != nil && task.execution.Binding.HasChecks() && task.executionVerified && c.VerificationStatus != "failed" {
			c.VerificationStatus = "passed"
		}
	}
	if res.Status != agentbridge.ResultCompleted && c.VerificationStatus == "pending" {
		c.VerificationStatus = "not_run"
	}
	if task != nil && task.sessionBindingErr != nil {
		c.DiagnosticCode, c.NextAction = "session_scope_conflict", "review_session"
	}
	return &c
}

func executionObservation(task *runningTask) execution.Observation {
	o := execution.Observation{AssignmentID: task.taskID, RuntimeID: task.runtimeID, Phase: "preparing"}
	if task.request != nil {
		if ids, err := workspacePrepareIDs(task.request); err == nil {
			o.WorkspaceID, o.TaskID, o.RunID = ids.workspaceID, ids.logicalTaskID, ids.runID
		}
		o.Directory, o.Phase = task.request.Cwd, "running"
		if task.request.HostExecutionBinding != nil {
			o.SourceDirectory = task.request.HostExecutionBinding.Repository
		}
	}
	if task.verificationPending {
		o.Phase = "verifying"
	}
	if task.cancelCause != nil {
		o.Phase = "cancelling"
	}
	if task.terminalResult != nil {
		o.Phase = "reporting"
	}
	return o
}
