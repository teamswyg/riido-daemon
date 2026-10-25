package contextcli

import "fmt"

// ExitCode is the process exit status of a work-context command.
type ExitCode int

const (
	ExitOK           ExitCode = 0
	ExitUsage        ExitCode = 2
	ExitAuthRequired ExitCode = 3
	ExitPermission   ExitCode = 4
	ExitNotFound     ExitCode = 5
	ExitUnavailable  ExitCode = 6
	ExitRateLimited  ExitCode = 7
	ExitUnsupported  ExitCode = 8
	ExitIncomplete   ExitCode = 9
	ExitConflict     ExitCode = 10
)

// ErrorCode is the stable machine-readable failure code in the envelope.
type ErrorCode string

const (
	CodeUsage                  ErrorCode = "USAGE"
	CodeSecretInArgv           ErrorCode = "SECRET_IN_ARGV"
	CodeAuthRequired           ErrorCode = "AUTH_REQUIRED"
	CodePermissionDenied       ErrorCode = "PERMISSION_DENIED"
	CodeOutOfRunScope          ErrorCode = "OUT_OF_RUN_SCOPE"
	CodeNotFound               ErrorCode = "NOT_FOUND"
	CodeSnapshotExpired        ErrorCode = "SNAPSHOT_EXPIRED"
	CodeUnavailable            ErrorCode = "UNAVAILABLE"
	CodeDaemonUnreachable      ErrorCode = "DAEMON_UNREACHABLE"
	CodeRateLimited            ErrorCode = "RATE_LIMITED"
	CodeUnsupported            ErrorCode = "UNSUPPORTED"
	CodeContextReadUnsupported ErrorCode = "CONTEXT_READ_UNSUPPORTED"
	CodeUserModeUnavailable    ErrorCode = "USER_MODE_UNAVAILABLE"
	CodeIncomplete             ErrorCode = "INCOMPLETE"
	CodeConflict               ErrorCode = "CONFLICT"
	CodeLeaseLost              ErrorCode = "LEASE_LOST"
)

// ExitCode maps an error code to the documented process exit status.
func (c ErrorCode) ExitCode() ExitCode {
	switch c {
	case CodeUsage, CodeSecretInArgv:
		return ExitUsage
	case CodeAuthRequired:
		return ExitAuthRequired
	case CodePermissionDenied, CodeOutOfRunScope:
		return ExitPermission
	case CodeNotFound, CodeSnapshotExpired:
		return ExitNotFound
	case CodeUnavailable, CodeDaemonUnreachable:
		return ExitUnavailable
	case CodeRateLimited:
		return ExitRateLimited
	case CodeUnsupported, CodeContextReadUnsupported, CodeUserModeUnavailable:
		return ExitUnsupported
	case CodeIncomplete:
		return ExitIncomplete
	case CodeConflict, CodeLeaseLost:
		return ExitConflict
	default:
		return ExitUnsupported
	}
}

// CLIError is a work-context failure with its envelope fields.
type CLIError struct {
	Code       ErrorCode
	Message    string
	NextAction string
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
	switch e.ExitCode() {
	case ExitUnavailable, ExitRateLimited:
		return true
	default:
		return false
	}
}
