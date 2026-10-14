package hostintegration

import (
	"testing"
	"time"
)

func TestConsentLedgerStateUsesLatestDecision(t *testing.T) {
	granted := consentRecord(ConsentProviderExecute, ConsentGranted)
	granted.Provider = "codex"
	revoked := consentRecord(ConsentProviderExecute, ConsentRevoked)
	revoked.Provider = "codex"
	revoked.RecordedAt = revoked.RecordedAt.Add(time.Minute)

	ledger, err := NewConsentLedger(granted, revoked)
	if err != nil {
		t.Fatal(err)
	}

	state := ledger.State()
	if state.ProviderExecutionAllowed("codex") {
		t.Fatal("latest revoked provider consent should win")
	}
	if got := len(ledger.Records()); got != 2 {
		t.Fatalf("records length = %d, want 2", got)
	}
}

func TestConsentLedgerStateTracksIndependentConsentKinds(t *testing.T) {
	ledger, err := NewConsentLedger(independentConsentKindRecords()...)
	if err != nil {
		t.Fatal(err)
	}

	state := ledger.State()
	assertIndependentConsentKinds(t, state)
}
