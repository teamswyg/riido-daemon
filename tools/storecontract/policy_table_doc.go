package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func applyPolicyTableOptions(
	repoRoot string,
	loaded contract,
	opts runOptions,
	result *checkResult,
) []string {
	if !opts.policyTableEnabled() {
		return nil
	}
	path := opts.PolicyTablePath
	if path == "" {
		path = defaultPolicyTablePath
	}
	rows, problems := buildPolicyTable(loaded.Channels)
	result.PolicyTablePath = path
	result.PolicyTableRows = rows
	if len(problems) > 0 {
		return problems
	}
	doc := renderPolicyTableDoc(loaded, rows)
	absPath := resolvePath(repoRoot, path)
	if opts.WritePolicyTable {
		return writePolicyTableDoc(absPath, doc)
	}
	if opts.CheckPolicyTable {
		return comparePolicyTableDoc(absPath, doc)
	}
	return nil
}

func writePolicyTableDoc(path, doc string) []string {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return []string{fmt.Sprintf("create policy table dir: %v", err)}
	}
	if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
		return []string{fmt.Sprintf("write policy table: %v", err)}
	}
	return nil
}

func comparePolicyTableDoc(path, expected string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("read policy table: %v", err)}
	}
	if string(data) != expected {
		return []string{"store channel policy table is stale; run tools/storecontract -write-policy-table"}
	}
	return nil
}
