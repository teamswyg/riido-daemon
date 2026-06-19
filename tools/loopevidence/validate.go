package main

import (
	"fmt"
	"os"
)

var requiredPhases = []string{"observe", "hypothesis", "execute", "evaluate", "retrospective"}

func validate(root string, m manifest) []string {
	var problems []string
	problems = append(problems, validateHeader(m)...)
	seen := map[string]bool{}
	for _, item := range m.Loops {
		problems = append(problems, validateLoop(root, item, seen)...)
	}
	for _, item := range m.OpenGaps {
		problems = append(problems, validateGap(item)...)
	}
	return problems
}

func validateHeader(m manifest) []string {
	var problems []string
	if m.SchemaVersion != "riido-loop-evidence.v1" {
		problems = append(problems, "schema_version must be riido-loop-evidence.v1")
	}
	for _, value := range []string{m.ID, m.Title, m.GeneratedDoc} {
		if value == "" {
			problems = append(problems, "id, title, and generated_doc are required")
		}
	}
	if fmt.Sprint(m.RequiredPhases) != fmt.Sprint(requiredPhases) {
		problems = append(problems, "required_phases must match the loop vocabulary")
	}
	return problems
}

func validatePath(root, owner, path string) []string {
	if path == "" || isCommandRef(path) {
		return nil
	}
	if _, err := os.Stat(resolvePath(root, path)); err != nil {
		return []string{fmt.Sprintf("%s references missing artifact %q", owner, path)}
	}
	return nil
}
