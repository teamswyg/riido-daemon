package main

import (
	"strings"
	"testing"
)

func TestPolicyTableIsGeneratedFromPolicyFunction(t *testing.T) {
	loaded := validContract()
	rows, problems := buildPolicyTable(loaded.Channels)
	if len(problems) > 0 {
		t.Fatalf("policy table problems: %v", problems)
	}
	doc := renderPolicyTableDoc(loaded, rows)
	for _, want := range []string{
		"Channel status | preferred-first | requires-redesign | preferred-first | store-review-ready",
		"Provider CLI bundling | forbidden | forbidden | forbidden | forbidden",
		"Provider CLI user-selected path | allowed | requires os-grant + store-review | allowed | allowed",
		"Background helper | requires explicit-consent | requires explicit-consent + store-review",
		"Windows service install | not applicable | not applicable | discouraged | forbidden",
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("generated policy table missing %q\n%s", want, doc)
		}
	}
}

func TestPolicyTableCheckRejectsStaleDoc(t *testing.T) {
	root := t.TempDir()
	writeRequiredDocs(t, root)
	writeContract(t, root, validContract())
	writeFile(t, resolvePath(root, defaultPolicyTablePath), "# stale\n")

	result, err := runWithOptions(root, "packaging/store/riido_daemon_store_distribution.riido.json", runOptions{
		PolicyTablePath:  defaultPolicyTablePath,
		CheckPolicyTable: true,
	})
	if err == nil {
		t.Fatalf("expected stale policy table error")
	}
	if !hasError(result.Errors, "store channel policy table is stale; run tools/storecontract -write-policy-table") {
		t.Fatalf("expected stale policy table error, got %v", result.Errors)
	}
}
