package hostintegration

// ConsentKind is a C11 permission surface that requires explicit user intent.
type ConsentKind string

const (
	ConsentBackgroundHelper ConsentKind = "background-helper"
	ConsentProviderExecute  ConsentKind = "provider-execute"
	ConsentWorkspaceAccess  ConsentKind = "workspace-access"
	ConsentTelemetrySync    ConsentKind = "telemetry-sync"
	ConsentReviewDemoMode   ConsentKind = "review-demo-mode"
)

// ConsentDecision is the append-only ledger action.
type ConsentDecision string

const (
	ConsentGranted ConsentDecision = "granted"
	ConsentRevoked ConsentDecision = "revoked"
)
