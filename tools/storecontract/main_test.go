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
