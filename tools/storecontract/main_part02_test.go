package main

import (
	"testing"
)

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
