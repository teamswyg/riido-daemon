package hostintegration

import (
	"errors"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

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
