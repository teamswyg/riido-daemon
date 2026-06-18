package hostintegration

import "testing"

func independentConsentKindRecords() []ConsentRecord {
	background := consentRecord(ConsentBackgroundHelper, ConsentGranted)
	telemetry := consentRecord(ConsentTelemetrySync, ConsentRevoked)
	demo := consentRecord(ConsentReviewDemoMode, ConsentGranted)
	provider := consentRecord(ConsentProviderExecute, ConsentGranted)
	provider.Provider = "claude"
	workspace := consentRecord(ConsentWorkspaceAccess, ConsentGranted)
	workspace.WorkspaceID = "workspace-1"
	return []ConsentRecord{background, telemetry, demo, provider, workspace}
}

func assertIndependentConsentKinds(t *testing.T, state ConsentState) {
	t.Helper()
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
