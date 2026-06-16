package taskvalidation

import (
	"errors"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/validation"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

type Request struct {
	TaskID         string
	Command        string
	Workdir        string
	Timeout        time.Duration
	Actor          string
	Source         string
	Summary        string
	Provider       string
	DecisionLLM    string
	ApprovalID     string
	CommandID      string
	ValidationGate string
}

type Result struct {
	TaskDB            taskdb.TaskDB
	Task              taskdb.TaskRecord
	Validation        validation.CommandResult
	Evidence          taskdb.TaskEvidenceRecord
	Receipt           taskdb.TaskCommandReceiptRecord
	Transition        *taskdb.TaskTransitionRecord
	TransitionReceipt *taskdb.TaskCommandReceiptRecord
	Provider          string
	CommandID         string
}

func (r Request) normalized() Request {
	r.TaskID = strings.TrimSpace(r.TaskID)
	r.Command = strings.TrimSpace(r.Command)
	r.Actor = textutil.Default(r.Actor, "daemon")
	r.Source = textutil.Default(r.Source, "riido-cli")
	r.Provider = strings.TrimSpace(r.Provider)
	r.DecisionLLM = strings.TrimSpace(r.DecisionLLM)
	r.ApprovalID = strings.TrimSpace(r.ApprovalID)
	r.CommandID = strings.TrimSpace(r.CommandID)
	return r
}

func (r Request) validate() error {
	if r.TaskID == "" {
		return errors.New("task_id is required")
	}
	if r.Command == "" {
		return errors.New("command is required")
	}
	if r.ApprovalID == "" {
		return errors.New("approval_id is required before validation command execution")
	}
	if r.Timeout < 0 {
		return errors.New("timeout must not be negative")
	}
	return nil
}
