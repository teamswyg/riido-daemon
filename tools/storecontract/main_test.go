package main

import (
	"os"
	"path/filepath"
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
