package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

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
