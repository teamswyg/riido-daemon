package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestWritePolicyTableDocReportsParentCreationFailure(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "not-dir"), "file")

	problems := writePolicyTableDoc(filepath.Join(root, "not-dir", "policy.md"), "doc")
	if len(problems) != 1 || !strings.Contains(problems[0], "create policy table dir:") {
		t.Fatalf("expected policy table dir problem, got %v", problems)
	}
}

func TestWriteJSONReportsParentCreationFailure(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "not-dir"), "file")

	err := writeJSON(filepath.Join(root, "not-dir", "result.json"), checkResult{})
	if err == nil || !strings.Contains(err.Error(), "create output dir:") {
		t.Fatalf("expected output dir problem, got %v", err)
	}
}

func TestDecodeJSONFileRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "contract.json")
	writeFile(t, path, `{"schema_version":"x","unknown":true}`)

	var loaded contract
	err := decodeJSONFile(path, &loaded)
	if err == nil || !strings.Contains(err.Error(), `unknown field "unknown"`) {
		t.Fatalf("expected unknown field decode problem, got %v", err)
	}
}

func TestValidateNoticeTermShapeCoversEmptyValidAndBlankTerms(t *testing.T) {
	if got := validateNoticeTermShape(nil); !hasError(got, "required_notice_terms must not be empty") {
		t.Fatalf("expected empty notice terms problem, got %v", got)
	}
	if got := validateNoticeTermShape([]string{"term"}); len(got) != 0 {
		t.Fatalf("valid notice terms should pass: %v", got)
	}
	if got := validateNoticeTermShape([]string{" "}); !hasError(got, "required_notice_terms must not include empty terms") {
		t.Fatalf("expected blank notice term problem, got %v", got)
	}
}
