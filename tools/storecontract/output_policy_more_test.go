package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteJSONCreatesParentAndTerminatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "result.json")
	err := writeJSON(path, checkResult{SchemaVersion: "test.v1", Status: "passed"})
	if err != nil {
		t.Fatalf("writeJSON: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	if !strings.HasSuffix(string(data), "\n") || !strings.Contains(string(data), `"status": "passed"`) {
		t.Fatalf("unexpected JSON output: %s", data)
	}
}

func TestPolicyTableOptionsWriteThenCheckFreshDoc(t *testing.T) {
	root := t.TempDir()
	loaded := validContract()
	var written checkResult
	problems := applyPolicyTableOptions(root, loaded, runOptions{WritePolicyTable: true}, &written)
	if len(problems) != 0 {
		t.Fatalf("write policy table problems: %v", problems)
	}
	if written.PolicyTablePath != defaultPolicyTablePath || len(written.PolicyTableRows) == 0 {
		t.Fatalf("policy table metadata missing: %#v", written)
	}
	if _, err := os.Stat(resolvePath(root, defaultPolicyTablePath)); err != nil {
		t.Fatalf("policy table not written: %v", err)
	}
	var checked checkResult
	problems = applyPolicyTableOptions(root, loaded, runOptions{CheckPolicyTable: true}, &checked)
	if len(problems) != 0 {
		t.Fatalf("fresh policy table should pass: %v", problems)
	}
}

func TestPolicyTableCompareReportsMissingDoc(t *testing.T) {
	problems := comparePolicyTableDoc(filepath.Join(t.TempDir(), "missing.md"), "expected")
	if len(problems) != 1 || !strings.Contains(problems[0], "read policy table") {
		t.Fatalf("expected missing policy table problem, got %v", problems)
	}
}

func TestResolvePathKeepsAbsolutePaths(t *testing.T) {
	absolute := filepath.Join(t.TempDir(), "doc.md")
	if got := resolvePath("ignored", absolute); got != absolute {
		t.Fatalf("absolute path changed: %q", got)
	}
	if got := resolvePath("/repo", "doc.md"); got != filepath.Join("/repo", "doc.md") {
		t.Fatalf("relative path not rooted: %q", got)
	}
}
