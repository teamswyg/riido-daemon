package hostintegration

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/provider/capability"
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

func TestExternalToolRecordStoreAutoDetectedRequiresConfirmation(t *testing.T) {
	record := validExternalToolRecord()
	record.Provenance = ToolProvenanceAutoDetected

	if !record.RequiresExecutionConfirmation(DistributionChannelMacAppStore) {
		t.Fatal("auto-detected CLI should require confirmation in mac app store channel")
	}
	if !record.RequiresExecutionConfirmation(DistributionChannelMSIXStore) {
		t.Fatal("auto-detected CLI should require confirmation in msix store channel")
	}
	if record.RequiresExecutionConfirmation(DistributionChannelDeveloperID) {
		t.Fatal("developer-id channel should not use the store confirmation rule")
	}
}

func TestExternalToolRecordServerFacingStatusStripsPrivatePath(t *testing.T) {
	record := validExternalToolRecord()

	status, err := record.ServerFacingStatus(DistributionChannelMacAppStore, "1.2.3")
	if err != nil {
		t.Fatalf("server status failed: %v", err)
	}

	if status.DistributionChannel != DistributionChannelMacAppStore ||
		status.AppVersion != "1.2.3" ||
		status.ProviderKind != record.Provider ||
		!status.ProviderAvailable ||
		status.ProviderLoginStatus != ToolLoginLoggedIn {
		t.Fatalf("unexpected status: %+v", status)
	}

	statusType := reflect.TypeOf(status)
	for _, forbidden := range []string{"ExecutablePath", "WorkspaceRootPath", "Token", "APIKey"} {
		if _, ok := statusType.FieldByName(forbidden); ok {
			t.Fatalf("server-facing status leaked forbidden field %s", forbidden)
		}
	}
}

func TestExternalToolRegistryPreservesStrongestProvenance(t *testing.T) {
	weak := validExternalToolRecord()
	weak.Provenance = ToolProvenanceAutoDetected
	weak.ExecutablePath = "/usr/local/bin/codex"

	strong := validExternalToolRecord()
	strong.Provenance = ToolProvenanceUserSelected
	strong.ExecutablePath = "/Applications/Codex.app/Contents/MacOS/codex"

	registry, err := NewExternalToolRegistry(strong)
	if err != nil {
		t.Fatalf("registry create failed: %v", err)
	}

	effective, accepted, err := registry.Register(weak)
	if err != nil {
		t.Fatalf("register weak failed: %v", err)
	}
	if accepted {
		t.Fatal("weaker provenance should not replace user-selected path")
	}
	if effective.ExecutablePath != strong.ExecutablePath {
		t.Fatalf("effective path changed: %+v", effective)
	}
}

func TestExternalToolRegistryAllowsStrongerProvenanceToReplace(t *testing.T) {
	weak := validExternalToolRecord()
	weak.Provenance = ToolProvenanceAutoDetected
	weak.ExecutablePath = "/usr/local/bin/codex"

	strong := validExternalToolRecord()
	strong.Provenance = ToolProvenanceEnvOverride
	strong.ExecutablePath = "/opt/homebrew/bin/codex"

	registry, err := NewExternalToolRegistry(weak)
	if err != nil {
		t.Fatalf("registry create failed: %v", err)
	}

	effective, accepted, err := registry.Register(strong)
	if err != nil {
		t.Fatalf("register strong failed: %v", err)
	}
	if !accepted {
		t.Fatal("stronger provenance should replace auto-detected path")
	}
	if effective.ExecutablePath != strong.ExecutablePath {
		t.Fatalf("effective path = %q, want %q", effective.ExecutablePath, strong.ExecutablePath)
	}
}

func TestExternalToolRegistryRecordsAreDeterministic(t *testing.T) {
	codex := validExternalToolRecord()
	codex.Provider = "codex"

	claude := validExternalToolRecord()
	claude.Provider = "claude"

	registry, err := NewExternalToolRegistry(codex, claude)
	if err != nil {
		t.Fatalf("registry create failed: %v", err)
	}

	records := registry.Records()
	if got, want := []capability.ProviderKind{records[0].Provider, records[1].Provider}, []capability.ProviderKind{"claude", "codex"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("records order = %v, want %v", got, want)
	}
}

func validExternalToolRecord() ExternalToolRecord {
	return ExternalToolRecord{
		Provider:            "codex",
		ExecutablePath:      "/usr/local/bin/codex",
		Provenance:          ToolProvenanceUserSelected,
		DetectedVersion:     "0.1.0",
		LoginStatus:         ToolLoginLoggedIn,
		CompatibilityStatus: capability.CompatSupported,
		LastVerifiedAt:      time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
	}
}
