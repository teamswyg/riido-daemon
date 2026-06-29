package main

import (
	"fmt"
	"os"
)

func validateLoopRegistry(root string, entries []loopRegistryEntry) []string {
	var problems []string
	seen := map[string]bool{}
	for _, entry := range entries {
		problems = append(problems, validateLoopRegistryEntry(root, entry, seen)...)
		seen[entry.ID] = true
	}
	return problems
}

func validateLoopRegistryEntry(root string, e loopRegistryEntry, seen map[string]bool) []string {
	var problems []string
	if e.ID == "" || e.LoopSource == "" || e.ExpiresAfter == "" {
		problems = append(problems, "loop registry id, loop_source, and expires_after are required")
	}
	if seen[e.ID] {
		problems = append(problems, fmt.Sprintf("duplicate loop registry id %q", e.ID))
	}
	if len(e.Observes) == 0 || len(e.Verifies) == 0 || len(e.Evidence) == 0 || len(e.FailsWhen) == 0 {
		problems = append(problems, fmt.Sprintf("loop registry %q must declare observes, verifies, evidence, and fails_when", e.ID))
	}
	if e.LoopSource != "" {
		if _, err := os.Stat(resolvePath(root, e.LoopSource)); err != nil {
			problems = append(problems, fmt.Sprintf("loop registry %q missing loop_source %q", e.ID, e.LoopSource))
		}
	}
	return problems
}
