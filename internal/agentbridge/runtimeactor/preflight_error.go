package runtimeactor

import (
	"errors"
	"syscall"
)

// PreflightError is emitted only before a provider process was started.
// Context errors from the public Submit wait are deliberately not promoted:
// the caller may not know whether the process crossed the dispatch boundary.
type PreflightError struct {
	Reason     string
	Retryable  bool
	Switchable bool
	Cause      error
}

func (e *PreflightError) Error() string { return "provider preflight: " + e.Reason }
func (e *PreflightError) Unwrap() error { return e.Cause }

func preflightError(reason string, retryable, switchable bool, err error) error {
	return &PreflightError{Reason: reason, Retryable: retryable, Switchable: switchable, Cause: err}
}

func transientSpawnError(err error) bool { return errors.Is(err, syscall.EAGAIN) }

// UnavailablePreflight classifies typed probe outcomes, never raw error text.
func UnavailablePreflight(capView Capability) error {
	retryable := capView.DiagnosticCode == "probe-failed" || capView.DiagnosticCode == "auth-probe-failed" || capView.DiagnosticCode == "version-probe-failed" || capView.DiagnosticCode == "capability-probe-failed"
	switchable := capView.DiagnosticCode == "executable-missing" || capView.DiagnosticCode == "version-unsupported"
	return preflightError("provider_unavailable", retryable, switchable, ErrProviderUnavailable)
}
