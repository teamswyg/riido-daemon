package hostintegration

import (
	"strings"
	"testing"
)

func TestExternalToolRecordValidate(t *testing.T) {
	record := validExternalToolRecord()
	if err := record.Validate(); err != nil {
		t.Fatalf("valid record rejected: %v", err)
	}
}

func TestExternalToolRecordValidateRejectsMissingRequiredFields(t *testing.T) {
	record := ExternalToolRecord{}
	err := record.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}

	for _, want := range []string{
		"provider is required",
		"executable path is required",
		"unknown provenance",
		"unknown login status",
		"unknown compatibility status",
		"last verified time is required",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("validation error %q missing %q", err, want)
		}
	}
}

func TestExternalToolRecordLoginRequiredIsValidButUnavailable(t *testing.T) {
	record := validExternalToolRecord()
	record.LoginStatus = ToolLoginRequired

	if err := record.Validate(); err != nil {
		t.Fatalf("login-required should be a valid registry row: %v", err)
	}
	if record.ProviderAvailable() {
		t.Fatal("login-required provider should not be routable")
	}
}
