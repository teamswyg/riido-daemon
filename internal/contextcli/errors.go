package contextcli

import (
	"fmt"

	contract "github.com/teamswyg/riido-contracts/contextcli"
)

// Exit and error codes, and their mapping, are owned by riido-contracts.
type (
	ExitCode  = contract.ExitCode
	ErrorCode = contract.ErrorCode
)

const (
	ExitOK           = contract.ExitOK
	ExitUsage        = contract.ExitUsage
	ExitAuthRequired = contract.ExitAuthRequired
	ExitPermission   = contract.ExitPermission
	ExitNotFound     = contract.ExitNotFound
	ExitUnavailable  = contract.ExitUnavailable
	ExitRateLimited  = contract.ExitRateLimited
	ExitUnsupported  = contract.ExitUnsupported
	ExitIncomplete   = contract.ExitIncomplete
	ExitConflict     = contract.ExitConflict
)

const (
	CodeUsage                   = contract.CodeUsage
	CodeSecretInArgv            = contract.CodeSecretInArgv
	CodeAuthRequired            = contract.CodeAuthRequired
	CodePermissionDenied        = contract.CodePermissionDenied
	CodeOutOfRunScope           = contract.CodeOutOfRunScope
	CodeNotFound                = contract.CodeNotFound
	CodeSnapshotExpired         = contract.CodeSnapshotExpired
	CodeUnavailable             = contract.CodeUnavailable
	CodeDaemonUnreachable       = contract.CodeDaemonUnreachable
	CodeRateLimited             = contract.CodeRateLimited
	CodeUnsupported             = contract.CodeUnsupported
	CodeContextReadUnsupported  = contract.CodeContextReadUnsupported
	CodeUserModeUnavailable     = contract.CodeUserModeUnavailable
	CodeControlPlaneUnsupported = contract.CodeControlPlaneUnsupported
	CodeIncomplete              = contract.CodeIncomplete
	CodeConflict                = contract.CodeConflict
	CodeLeaseLost               = contract.CodeLeaseLost
)

// CLIError is a work-context failure with its envelope fields.
type CLIError struct {
	Code       ErrorCode
	Message    string
	NextAction string
	// NoRetry prevents retries for limits that cannot recover in this execution.
	NoRetry bool
}

// Fail builds a CLIError.
func Fail(code ErrorCode, message, nextAction string) *CLIError {
	return &CLIError{Code: code, Message: message, NextAction: nextAction}
}

// Failf builds a CLIError with a formatted message and no next action.
func Failf(code ErrorCode, format string, args ...any) *CLIError {
	return &CLIError{Code: code, Message: fmt.Sprintf(format, args...)}
}

func (e *CLIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ExitCode returns the process exit status for e.
func (e *CLIError) ExitCode() ExitCode {
	return e.Code.ExitCode()
}

// Retryable reports whether retrying the same request may succeed.
func (e *CLIError) Retryable() bool {
	if e.NoRetry {
		return false
	}
	switch e.ExitCode() {
	case ExitUnavailable, ExitRateLimited:
		return true
	default:
		return false
	}
}
