package supervisor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/execution"
)

type taskVerificationMsg struct {
	task     *runningTask
	result   agentbridge.Result
	evidence execution.Evidence
}

func (a *Actor) beginExecutionVerification(task *runningTask, res agentbridge.Result) {
	if task.verificationPending || task.terminalResult != nil {
		return
	}
	task.verificationPending = true
	// Verification owns only immutable workspace references and its cancellation
	// context. The supervisor remains the sole owner of task lifecycle state.
	snapshot := &runningTask{ctx: task.ctx, execution: task.execution, workspace: task.workspace}
	go func() {
		result, evidence := verifyExecutionResult(task.ctx, snapshot, res)
		select {
		case a.mailbox <- envelope{taskVerification: &taskVerificationMsg{task: task, result: result, evidence: evidence}}:
		case <-a.stoppedCh:
		}
	}()
}

func verifyExecutionResult(ctx context.Context, task *runningTask, res agentbridge.Result) (agentbridge.Result, execution.Evidence) {
	if task.ctx != nil {
		ctx = task.ctx
	}
	evidence := task.execution.Verify(ctx)
	data, err := json.MarshalIndent(evidence, "", "  ")
	if err == nil && task.workspace != nil {
		err = os.WriteFile(filepath.Join(task.workspace.Artifacts, "verification.json"), data, 0o600)
	}
	res.Workdir = task.execution.Directory
	res.FinishedAt = time.Now()
	switch {
	case ctx.Err() != nil:
		res.Status, res.Error = agentbridge.ResultCancelled, "Execution verification cancelled"
	case err != nil:
		res.Status, res.Error = agentbridge.ResultFailed, "Could not persist execution evidence: "+err.Error()
		evidence.ErrorKind = "evidence_persistence"
	case evidence.Error != "":
		res.Status, res.Error = agentbridge.ResultFailed, "Execution verification failed: "+evidence.Error
	default:
		res.Output += "\n\nVerified changed files: " + strings.Join(evidence.ChangedFiles, ", ") + fmt.Sprintf("; checks passed: %d", len(evidence.Checks))
	}
	return res, evidence
}

func (a *Actor) handleExecutionVerification(ctx context.Context, msg *taskVerificationMsg, inFlight map[string]*runningTask) {
	task := inFlight[msg.task.taskID]
	if task != msg.task || task.terminalResult != nil || !task.verificationPending {
		return
	}
	task.verificationPending = false
	res := msg.result
	if task.ctx.Err() != nil || task.cancelCause != nil {
		res.Status, res.Error = agentbridge.ResultCancelled, "Execution verification cancelled"
	} else if canContinueExecution(task, res, msg.evidence) {
		a.continueExecution(ctx, task, res, msg.evidence)
		return
	}
	task.executionVerified = true
	res = a.recordTerminalResult(ctx, task, res)
	a.finishTaskWithResult(ctx, inFlight, task, res)
}
