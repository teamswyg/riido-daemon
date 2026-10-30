package userauth

import (
	"context"
	"errors"
)

var (
	ErrPreviousGrantRevocationFailed = errors.New("new CLI session is active but previous grant revocation failed")
	ErrNotAuthenticated              = errors.New("CLI authentication is required")
	ErrReauthRequired                = errors.New("CLI authentication must be renewed")
	ErrStoreLocked                   = errors.New("credential storage requires user confirmation")
	ErrLoggedOut                     = errors.New("CLI authentication changed during the request")
	ErrLocalRequestCleanupFailed     = errors.New("protected local authentication requests could not be cleared")
)

// Error intentionally excludes provider messages, URLs, bodies and tokens.
type Error struct {
	Code       string
	StatusCode int
	Retryable  bool
}

func (e *Error) Error() string {
	switch e.Code {
	case "INVALID_CONFIG":
		return "invalid CLI authentication configuration"
	case "AUTH_CALLBACK":
		return "invalid CLI authorization response"
	case "AUTH_DENIED":
		return "CLI authorization was not granted"
	case "AUTH_TIMEOUT":
		return "CLI authorization timed out"
	case "AUTH_USED":
		return "CLI authorization response has already been consumed"
	case "REAUTH_REQUIRED":
		return ErrReauthRequired.Error()
	case "INVALID_RESPONSE":
		return "authentication server returned an invalid response"
	case "REQUEST_REJECTED":
		return "authentication server rejected the request"
	case "HTTP_ERROR":
		return "authentication server is temporarily unavailable"
	case "STORAGE_ERROR":
		return "protected CLI credential storage is unavailable"
	case "UNSAFE_REQUEST":
		return "request is outside the authenticated API origin"
	default:
		return "CLI authentication request failed"
	}
}

func (e *Error) Is(target error) bool {
	return e.Code == "REAUTH_REQUIRED" && target == ErrReauthRequired
}
func failure(code string) error { return &Error{Code: code} }

func storageError(err error) error {
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}
	if errors.Is(err, ErrNotAuthenticated) {
		return ErrNotAuthenticated
	}
	if errors.Is(err, ErrStoreLocked) {
		return ErrStoreLocked
	}
	return failure("STORAGE_ERROR")
}
