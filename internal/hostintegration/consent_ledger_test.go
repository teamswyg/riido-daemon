package hostintegration

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

func TestConsentRecordValidate(t *testing.T) {
	record := consentRecord(ConsentProviderExecute, ConsentGranted)
	record.Provider = "codex"

	if err := record.Validate(); err != nil {
		t.Fatalf("valid provider consent rejected: %v", err)
	}
}

func TestConsentRecordValidateRejectsMissingRequiredFields(t *testing.T) {
	record := ConsentRecord{}

	err := record.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	for _, want := range []string{
		"unknown consent kind",
		"unknown consent decision",
		"recorded time is required",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("validation error %q missing %q", err, want)
		}
	}
}

func TestConsentRecordValidateRequiresCorrectSubject(t *testing.T) {
	providerRecord := consentRecord(ConsentProviderExecute, ConsentGranted)
	if err := providerRecord.Validate(); err == nil {
		t.Fatal("expected provider execute consent without provider to fail")
	}

	workspaceRecord := consentRecord(ConsentWorkspaceAccess, ConsentGranted)
	workspaceRecord.Provider = "codex"
	if err := workspaceRecord.Validate(); err == nil {
		t.Fatal("expected workspace access consent with provider to fail")
	}

	backgroundRecord := consentRecord(ConsentBackgroundHelper, ConsentGranted)
	backgroundRecord.WorkspaceID = "workspace-1"
	if err := backgroundRecord.Validate(); err == nil {
		t.Fatal("expected global consent with workspace id to fail")
	}
}

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
	background := consentRecord(ConsentBackgroundHelper, ConsentGranted)
	telemetry := consentRecord(ConsentTelemetrySync, ConsentRevoked)
	demo := consentRecord(ConsentReviewDemoMode, ConsentGranted)
	provider := consentRecord(ConsentProviderExecute, ConsentGranted)
	provider.Provider = "claude"
	workspace := consentRecord(ConsentWorkspaceAccess, ConsentGranted)
	workspace.WorkspaceID = "workspace-1"

	ledger, err := NewConsentLedger(background, telemetry, demo, provider, workspace)
	if err != nil {
		t.Fatal(err)
	}

	state := ledger.State()
	if !state.BackgroundHelper {
		t.Fatal("background helper should be granted")
	}
	if state.TelemetrySync {
		t.Fatal("telemetry sync should be revoked")
	}
	if !state.ReviewDemoMode {
		t.Fatal("review demo mode should be granted")
	}
	if !state.ProviderExecutionAllowed("claude") {
		t.Fatal("provider execute should be granted")
	}
	if !state.WorkspaceAccessAllowed("workspace-1") {
		t.Fatal("workspace access should be granted")
	}
}

func TestConsentStateGrantedSubjectsAreDeterministic(t *testing.T) {
	state := ConsentState{
		ProviderExecute: map[capability.ProviderKind]bool{
			"codex":  true,
			"claude": true,
			"cursor": false,
		},
		WorkspaceAccess: map[string]bool{
			"workspace-z": true,
			"workspace-a": true,
			"workspace-b": false,
		},
	}

	if got, want := state.GrantedProviders(), []capability.ProviderKind{"claude", "codex"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("providers = %v, want %v", got, want)
	}
	if got, want := state.GrantedWorkspaces(), []string{"workspace-a", "workspace-z"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("workspaces = %v, want %v", got, want)
	}
}

func consentRecord(kind ConsentKind, decision ConsentDecision) ConsentRecord {
	return ConsentRecord{
		Kind:       kind,
		Decision:   decision,
		Actor:      "user:tester",
		RecordedAt: time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
	}
}
