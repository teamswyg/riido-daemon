package hostintegration

import (
	"errors"
	"slices"
	"sort"
	"time"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

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

// ConsentRecord is one immutable user-intent fact. Provider and WorkspaceID
// are mutually exclusive subjects depending on ConsentKind.
type ConsentRecord struct {
	Kind        ConsentKind
	Decision    ConsentDecision
	Provider    capability.ProviderKind
	WorkspaceID string
	Actor       string
	Reason      string
	RecordedAt  time.Time
}

// ConsentState is the current view reduced from append-only records.
type ConsentState struct {
	BackgroundHelper bool
	TelemetrySync    bool
	ReviewDemoMode   bool
	ProviderExecute  map[capability.ProviderKind]bool
	WorkspaceAccess  map[string]bool
}

// ConsentLedger is an append-only in-memory view. Persistence adapters store
// records; the current state is always derived.
type ConsentLedger struct {
	records []ConsentRecord
}

// NewConsentLedger validates and appends records in order.
func NewConsentLedger(records ...ConsentRecord) (*ConsentLedger, error) {
	ledger := &ConsentLedger{}
	for _, record := range records {
		if err := ledger.Append(record); err != nil {
			return nil, err
		}
	}
	return ledger, nil
}

// Append validates and records a new consent fact.
func (l *ConsentLedger) Append(record ConsentRecord) error {
	if l == nil {
		return errors.New("consent ledger is nil")
	}
	if err := record.Validate(); err != nil {
		return err
	}
	l.records = append(l.records, record)
	return nil
}

// Records returns a copy of the append-only facts in insertion order.
func (l *ConsentLedger) Records() []ConsentRecord {
	if l == nil || len(l.records) == 0 {
		return nil
	}
	out := make([]ConsentRecord, len(l.records))
	copy(out, l.records)
	return out
}

// State reduces the ledger to the current consent view.
func (l *ConsentLedger) State() ConsentState {
	state := ConsentState{
		ProviderExecute: make(map[capability.ProviderKind]bool),
		WorkspaceAccess: make(map[string]bool),
	}
	if l == nil {
		return state
	}
	for _, record := range l.records {
		granted := record.Decision == ConsentGranted
		switch record.Kind {
		case ConsentBackgroundHelper:
			state.BackgroundHelper = granted
		case ConsentTelemetrySync:
			state.TelemetrySync = granted
		case ConsentReviewDemoMode:
			state.ReviewDemoMode = granted
		case ConsentProviderExecute:
			state.ProviderExecute[record.Provider] = granted
		case ConsentWorkspaceAccess:
			state.WorkspaceAccess[record.WorkspaceID] = granted
		}
	}
	return state
}

// ProviderExecutionAllowed reports whether a provider can be executed under
// the latest consent view.
func (s ConsentState) ProviderExecutionAllowed(provider capability.ProviderKind) bool {
	return s.ProviderExecute[provider]
}

// WorkspaceAccessAllowed reports whether a workspace root grant is active.
func (s ConsentState) WorkspaceAccessAllowed(workspaceID string) bool {
	return s.WorkspaceAccess[workspaceID]
}

// GrantedProviders returns active provider grants in deterministic order.
func (s ConsentState) GrantedProviders() []capability.ProviderKind {
	providers := make([]capability.ProviderKind, 0, len(s.ProviderExecute))
	for provider, granted := range s.ProviderExecute {
		if granted {
			providers = append(providers, provider)
		}
	}
	slices.Sort(providers)
	return providers
}

// GrantedWorkspaces returns active workspace grants in deterministic order.
func (s ConsentState) GrantedWorkspaces() []string {
	workspaces := make([]string, 0, len(s.WorkspaceAccess))
	for workspaceID, granted := range s.WorkspaceAccess {
		if granted {
			workspaces = append(workspaces, workspaceID)
		}
	}
	sort.Strings(workspaces)
	return workspaces
}
