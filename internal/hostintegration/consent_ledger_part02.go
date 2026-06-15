package hostintegration

import (
	"errors"
	"fmt"
	"strings"
)

// Validate checks the consent record shape. It does not verify identity of the
// actor or whether a workspace path exists; those belong to adapters.
func (r ConsentRecord) Validate() error {
	var errs []error
	if !r.Kind.Valid() {
		errs = append(errs, fmt.Errorf("unknown consent kind %q", r.Kind))
	}
	if !r.Decision.Valid() {
		errs = append(errs, fmt.Errorf("unknown consent decision %q", r.Decision))
	}
	if r.RecordedAt.IsZero() {
		errs = append(errs, errors.New("recorded time is required"))
	}

	provider := strings.TrimSpace(string(r.Provider))
	workspaceID := strings.TrimSpace(r.WorkspaceID)
	switch r.Kind {
	case ConsentProviderExecute:
		if provider == "" {
			errs = append(errs, errors.New("provider execute consent requires provider"))
		}
		if workspaceID != "" {
			errs = append(errs, errors.New("provider execute consent must not include workspace id"))
		}
	case ConsentWorkspaceAccess:
		if workspaceID == "" {
			errs = append(errs, errors.New("workspace access consent requires workspace id"))
		}
		if provider != "" {
			errs = append(errs, errors.New("workspace access consent must not include provider"))
		}
	case ConsentBackgroundHelper, ConsentTelemetrySync, ConsentReviewDemoMode:
		if provider != "" || workspaceID != "" {
			errs = append(errs, fmt.Errorf("%s consent must not include provider or workspace id", r.Kind))
		}
	}
	return errors.Join(errs...)
}

// Valid reports whether kind is one of the SSOT-defined consent kinds.
func (kind ConsentKind) Valid() bool {
	switch kind {
	case ConsentBackgroundHelper,
		ConsentProviderExecute,
		ConsentWorkspaceAccess,
		ConsentTelemetrySync,
		ConsentReviewDemoMode:
		return true
	default:
		return false
	}
}

// Valid reports whether decision is one of the SSOT-defined decisions.
func (decision ConsentDecision) Valid() bool {
	switch decision {
	case ConsentGranted, ConsentRevoked:
		return true
	default:
		return false
	}
}
