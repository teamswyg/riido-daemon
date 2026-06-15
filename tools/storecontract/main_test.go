package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestRunAcceptsValidContract(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	writeContract(t, root, validContract())

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err != nil {
		t.Fatalf("run returned error: %v\nerrors=%v", err, result.Errors)
	}
	if result.Status != "passed" {
		t.Fatalf("expected passed, got %s", result.Status)
	}
}

func TestRunRejectsBundledProviderCLI(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	writeContract(t, root, validContract())
	writeFile(t, filepath.Join(root, "packaging/store/claude"), "binary")

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected bundled provider CLI error")
	}
	if len(result.Errors) == 0 {
		t.Fatalf("expected validation errors")
	}
}

func TestRunRejectsMissingRequiredDoc(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	writeContract(t, root, validContract())
	if err := os.Remove(filepath.Join(root, "NOTICE.md")); err != nil {
		t.Fatal(err)
	}

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected missing doc error")
	}
	if len(result.Errors) == 0 {
		t.Fatalf("expected validation errors")
	}
}

func TestRunRejectsMissingNoticeTerm(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	writeFile(t, filepath.Join(root, "NOTICE.md"), "# NOTICE\nNo vendored third-party code\n")
	writeContract(t, root, validContract())

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected missing NOTICE provenance term error")
	}
	if !hasError(result.Errors, `NOTICE.md must include "Modified Apache License, Version 2.0"`) {
		t.Fatalf("expected missing NOTICE term error, got %v", result.Errors)
	}
}

func TestRunRejectsDeveloperIDWithoutNotarization(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	contract := validContract()
	for i := range contract.Channels {
		if contract.Channels[i].ID == "developer-id" {
			contract.Channels[i].RequiredSurfaces = removeString(contract.Channels[i].RequiredSurfaces, "notarization")
		}
	}
	writeContract(t, root, contract)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected missing Developer ID notarization error")
	}
	if !hasError(result.Errors, `channel "developer-id" must require notarization`) {
		t.Fatalf("expected notarization error, got %v", result.Errors)
	}
}

func TestRunRejectsMSIXSideloadWithoutSignedPackage(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	contract := validContract()
	for i := range contract.Channels {
		if contract.Channels[i].ID == "msix-sideload" {
			contract.Channels[i].RequiredSurfaces = removeString(contract.Channels[i].RequiredSurfaces, "signed-msix-package")
		}
	}
	writeContract(t, root, contract)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected missing signed MSIX package error")
	}
	if !hasError(result.Errors, `channel "msix-sideload" must require signed-msix-package`) {
		t.Fatalf("expected signed package error, got %v", result.Errors)
	}
}

func TestRunRejectsMSIXStoreWithoutNamedPipeIPC(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	contract := validContract()
	for i := range contract.Channels {
		if contract.Channels[i].ID == "msix-store" {
			contract.Channels[i].RequiredSurfaces = removeString(contract.Channels[i].RequiredSurfaces, "named-pipe-local-ipc")
		}
	}
	writeContract(t, root, contract)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected missing named pipe IPC error")
	}
	if !hasError(result.Errors, `channel "msix-store" must require named-pipe-local-ipc`) {
		t.Fatalf("expected named pipe IPC error, got %v", result.Errors)
	}
}

func TestRunRejectsMacAppStoreWithoutSandboxReviewNotes(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	contract := validContract()
	for i := range contract.Channels {
		if contract.Channels[i].ID == "mac-app-store" {
			contract.Channels[i].RequiredSurfaces = removeString(contract.Channels[i].RequiredSurfaces, "app-sandbox-entitlement-review-notes")
		}
	}
	writeContract(t, root, contract)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected missing App Sandbox review note error")
	}
	if !hasError(result.Errors, `channel "mac-app-store" must require app-sandbox-entitlement-review-notes`) {
		t.Fatalf("expected App Sandbox review note error, got %v", result.Errors)
	}
}

func TestRunRejectsMacAppStoreWithoutHelperPurposeReviewNote(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	contract := validContract()
	for i := range contract.Channels {
		if contract.Channels[i].ID == "mac-app-store" {
			contract.Channels[i].RequiredSurfaces = removeString(contract.Channels[i].RequiredSurfaces, "helper-purpose-review-note")
		}
	}
	writeContract(t, root, contract)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected missing helper purpose review note error")
	}
	if !hasError(result.Errors, `channel "mac-app-store" must require helper-purpose-review-note`) {
		t.Fatalf("expected helper purpose review note error, got %v", result.Errors)
	}
}

func TestRunRejectsMacAppStoreWithoutSelfUpdaterBan(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	contract := validContract()
	for i := range contract.Channels {
		if contract.Channels[i].ID == "mac-app-store" {
			contract.Channels[i].ForbiddenSurfaces = removeString(contract.Channels[i].ForbiddenSurfaces, "self-updater")
		}
	}
	writeContract(t, root, contract)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected missing self-updater ban error")
	}
	if !hasError(result.Errors, `channel "mac-app-store" must forbid self-updater`) {
		t.Fatalf("expected self-updater ban error, got %v", result.Errors)
	}
}

func TestRunRejectsMacAppStoreWithoutLoginItemRuntimeRole(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	contract := validContract()
	for i := range contract.Channels {
		if contract.Channels[i].ID == "mac-app-store" {
			contract.Channels[i].RuntimeRole = "local-helper-broker"
		}
	}
	writeContract(t, root, contract)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected wrong mac-app-store runtime role error")
	}
	if !hasError(result.Errors, `channel "mac-app-store" runtime_role must be sandboxed-login-item-helper`) {
		t.Fatalf("expected sandboxed login item runtime role error, got %v", result.Errors)
	}
}

func TestRunRejectsMacAppStoreWithoutConsentAndReviewBackgroundRule(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	contract := validContract()
	for i := range contract.Channels {
		if contract.Channels[i].ID == "mac-app-store" {
			contract.Channels[i].BackgroundRule = "explicit-consent"
		}
	}
	writeContract(t, root, contract)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected wrong mac-app-store background rule error")
	}
	if !hasError(result.Errors, `channel "mac-app-store" background_rule must be explicit-consent-and-store-review`) {
		t.Fatalf("expected consent+review background rule error, got %v", result.Errors)
	}
}

func TestRunRejectsMacAppStoreWithoutAppStoreManagedUpdates(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	contract := validContract()
	for i := range contract.Channels {
		if contract.Channels[i].ID == "mac-app-store" {
			contract.Channels[i].UpdateMechanism = "self-managed"
		}
	}
	writeContract(t, root, contract)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected wrong mac-app-store update mechanism error")
	}
	if !hasError(result.Errors, `channel "mac-app-store" update_mechanism must be app-store-managed`) {
		t.Fatalf("expected app-store managed update mechanism error, got %v", result.Errors)
	}
}

func TestRunRejectsMSIXStoreWithoutReviewNotes(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	contract := validContract()
	for i := range contract.Channels {
		if contract.Channels[i].ID == "msix-store" {
			contract.Channels[i].RequiredSurfaces = removeString(contract.Channels[i].RequiredSurfaces, "runfulltrust-review-note")
		}
	}
	writeContract(t, root, contract)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected missing runFullTrust review note error")
	}
	if !hasError(result.Errors, `channel "msix-store" must require runfulltrust-review-note`) {
		t.Fatalf("expected runFullTrust review note error, got %v", result.Errors)
	}
}

func TestRunRejectsMSIXStoreWithoutFullTrustTrayRuntimeRole(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	contract := validContract()
	for i := range contract.Channels {
		if contract.Channels[i].ID == "msix-store" {
			contract.Channels[i].RuntimeRole = "msix-packaged-helper-broker"
		}
	}
	writeContract(t, root, contract)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected wrong msix-store runtime role error")
	}
	if !hasError(result.Errors, `channel "msix-store" runtime_role must be msix-packaged-full-trust-helper-tray`) {
		t.Fatalf("expected full-trust helper/tray runtime role error, got %v", result.Errors)
	}
}

func TestRunRejectsMSIXStoreWithoutConsentAndReviewBackgroundRule(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	contract := validContract()
	for i := range contract.Channels {
		if contract.Channels[i].ID == "msix-store" {
			contract.Channels[i].BackgroundRule = "explicit-consent"
		}
	}
	writeContract(t, root, contract)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected wrong msix-store background rule error")
	}
	if !hasError(result.Errors, `channel "msix-store" background_rule must be explicit-consent-and-store-review`) {
		t.Fatalf("expected consent+review background rule error, got %v", result.Errors)
	}
}

func TestRunRejectsMSIXStoreWithoutStoreManagedUpdates(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	contract := validContract()
	for i := range contract.Channels {
		if contract.Channels[i].ID == "msix-store" {
			contract.Channels[i].UpdateMechanism = "self-managed"
		}
	}
	writeContract(t, root, contract)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected wrong msix-store update mechanism error")
	}
	if !hasError(result.Errors, `channel "msix-store" update_mechanism must be store-managed`) {
		t.Fatalf("expected store-managed update mechanism error, got %v", result.Errors)
	}
}

func TestRunRejectsStoreChannelWithoutReviewDemoMode(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	contract := validContract()
	for i := range contract.Channels {
		if contract.Channels[i].ID == "mac-app-store" {
			contract.Channels[i].RequiredSurfaces = removeString(contract.Channels[i].RequiredSurfaces, "review-demo-mode")
		}
	}
	writeContract(t, root, contract)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected missing review demo mode error")
	}
	if !hasError(result.Errors, `channel "mac-app-store" must require review-demo-mode`) {
		t.Fatalf("expected review demo mode error, got %v", result.Errors)
	}
}

func TestRunRejectsStoreChannelWithoutReviewSubmissionSurface(t *testing.T) {
	tests := []struct {
		name    string
		channel string
		surface string
	}{
		{
			name:    "mac app store requires demo review account",
			channel: "mac-app-store",
			surface: "demo-review-account",
		},
		{
			name:    "mac app store requires privacy metadata allowlist",
			channel: "mac-app-store",
			surface: "privacy-metadata-allowlist",
		},
		{
			name:    "mac app store requires provider non bundling review note",
			channel: "mac-app-store",
			surface: "provider-non-bundling-review-note",
		},
		{
			name:    "microsoft store requires demo review account",
			channel: "msix-store",
			surface: "demo-review-account",
		},
		{
			name:    "microsoft store requires privacy metadata allowlist",
			channel: "msix-store",
			surface: "privacy-metadata-allowlist",
		},
		{
			name:    "microsoft store requires provider non bundling review note",
			channel: "msix-store",
			surface: "provider-non-bundling-review-note",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeRequiredDocs(t, root)
			contract := validContract()
			for i := range contract.Channels {
				if contract.Channels[i].ID == tt.channel {
					contract.Channels[i].RequiredSurfaces = removeString(contract.Channels[i].RequiredSurfaces, tt.surface)
				}
			}
			writeContract(t, root, contract)

			result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
			if err == nil {
				t.Fatalf("expected missing %s error", tt.surface)
			}
			wanted := `channel "` + tt.channel + `" must require ` + tt.surface
			if !hasError(result.Errors, wanted) {
				t.Fatalf("expected %s error, got %v", tt.surface, result.Errors)
			}
		})
	}
}

func validContract() contract {
	channels := []channel{
		{
			ID:                "developer-id",
			Platform:          "macos",
			Status:            "preferred-first",
			RuntimeRole:       "local-helper-broker",
			BackgroundRule:    "explicit-consent",
			LocalIPCTransport: "unix-socket",
			DataRoot:          "user-application-support",
			UpdateMechanism:   "self-managed",
			RequiredSurfaces:  []string{"developer-id-signing", "notarization", "user-consented-background-helper", "local-only-ipc"},
			ForbiddenSurfaces: []string{"bundled-provider-cli", "silent-provider-install", "external-tcp-listener", "arbitrary-home-scan"},
		},
		{
			ID:                "mac-app-store",
			Platform:          "macos",
			Status:            "requires-redesign",
			RuntimeRole:       "sandboxed-login-item-helper",
			BackgroundRule:    "explicit-consent-and-store-review",
			LocalIPCTransport: "unix-socket",
			DataRoot:          "app-group-or-sandbox-container",
			UpdateMechanism:   "app-store-managed",
			RequiredSurfaces:  []string{"app-sandbox", "app-group-or-container-ipc", "security-scoped-workspace-grant", "service-management-login-item-consent", "helper-purpose-review-note", "app-sandbox-entitlement-review-notes", "app-store-managed-updates", "privacy-policy", "review-demo-mode", "demo-review-account", "privacy-metadata-allowlist", "provider-non-bundling-review-note"},
			ForbiddenSurfaces: []string{"bundled-provider-cli", "silent-provider-install", "direct-launchagent-install", "self-updater", "external-tcp-listener", "arbitrary-home-scan", "third-party-installer", "shared-location-code-install", "standalone-code-download", "root-privilege-escalation"},
		},
		{
			ID:                "msix-sideload",
			Platform:          "windows",
			Status:            "preferred-first",
			RuntimeRole:       "msix-packaged-helper-broker",
			BackgroundRule:    "explicit-consent",
			LocalIPCTransport: "windows-named-pipe",
			DataRoot:          "windows-package-local-data",
			UpdateMechanism:   "self-managed",
			RequiredSurfaces:  []string{"signed-msix-package", "package-identity", "windows-desktop-target-device-family", "named-pipe-local-ipc", "package-local-data", "user-consented-background-helper"},
			ForbiddenSurfaces: []string{"bundled-provider-cli", "silent-provider-install", "windows-service-default", "external-tcp-listener", "arbitrary-home-scan"},
		},
		{
			ID:                "msix-store",
			Platform:          "windows",
			Status:            "requires-policy-gate",
			RuntimeRole:       "msix-packaged-full-trust-helper-tray",
			BackgroundRule:    "explicit-consent-and-store-review",
			LocalIPCTransport: "windows-named-pipe",
			DataRoot:          "windows-package-local-data",
			UpdateMechanism:   "store-managed",
			RequiredSurfaces:  []string{"package-identity", "windows-desktop-target-device-family", "named-pipe-local-ipc", "package-local-data", "runfulltrust-review-note", "store-managed-updates", "partner-center-review-notes", "review-demo-mode", "privacy-policy", "demo-review-account", "privacy-metadata-allowlist", "provider-non-bundling-review-note"},
			ForbiddenSurfaces: []string{"bundled-provider-cli", "silent-provider-install", "windows-service-default", "self-updater", "external-tcp-listener", "arbitrary-home-scan"},
		},
	}
	return contract{
		SchemaVersion:            contractSchemaVersion,
		Product:                  "riido_daemon",
		ProviderCLIBundling:      "forbidden",
		ExternalProviderCLINames: []string{"claude", "codex", "openclaw", "cursor-agent"},
		StoreArtifactRoots:       []string{"packaging/store"},
		RequiredDocs: []string{
			"docs/20-domain/distribution-host-integration.md",
			"docs/30-architecture/store-distribution.md",
			"NOTICE.md",
		},
		RequiredNoticeTerms: []string{
			"No source code from any third-party project is directly incorporated",
			"Modified Apache License, Version 2.0",
			"do not redistribute any vendor code or bundle provider CLI executables",
			"No vendored third-party code",
		},
		Channels: channels,
	}
}

func writeRequiredDocs(t *testing.T, root string) {
	t.Helper()
	writeFile(t, filepath.Join(root, "docs/20-domain/distribution-host-integration.md"), "# Distribution\n")
	writeFile(t, filepath.Join(root, "docs/30-architecture/store-distribution.md"), "# Store\n")
	writeFile(t, filepath.Join(root, "NOTICE.md"), "# NOTICE\nNo source code from any third-party project is directly incorporated\nModified Apache License, Version 2.0\ndo not redistribute any vendor code or bundle provider CLI executables\nNo vendored third-party code\n")
}

func writeContract(t *testing.T, root string, value contract) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "packaging/store/riido_daemon_store_distribution.riido.json"), string(data))
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func removeString(items []string, unwanted string) []string {
	var out []string
	for _, item := range items {
		if item != unwanted {
			out = append(out, item)
		}
	}
	return out
}

func hasError(errors []string, wanted string) bool {
	return slices.Contains(errors, wanted)
}
