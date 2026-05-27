package agentbridge

import "time"

// ResultStatus enumerates the terminal outcomes of a run.
//
// Terminal-status priority on conflict:
//
//	cancelled > timeout > failed > completed
//
// Process exit alone never sets a terminal status; it can only force failed
// when no provider result has been observed yet (reducer invariant 5).
type ResultStatus string

const (
	ResultCompleted ResultStatus = "completed"
	ResultFailed    ResultStatus = "failed"
	ResultCancelled ResultStatus = "cancelled"
	ResultTimeout   ResultStatus = "timeout"
	ResultAborted   ResultStatus = "aborted"
	ResultBlocked   ResultStatus = "blocked"
)

// Result is the terminal outcome of a run.
type Result struct {
	Status     ResultStatus
	Output     string
	Error      string
	SessionID  string
	Workdir    string
	Usage      Usage
	StartedAt  time.Time
	FinishedAt time.Time
}
