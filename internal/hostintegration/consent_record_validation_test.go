package hostintegration

import (
	"strings"
	"testing"
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
	for _, want := range missingConsentRecordFields() {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("validation error %q missing %q", err, want)
		}
	}
}

func missingConsentRecordFields() []string {
	return []string{
		"unknown consent kind",
		"unknown consent decision",
		"recorded time is required",
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
