package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var requiredPhases = []string{"observe", "hypothesis", "execute", "evaluate", "retrospective"}

func validate(root string, m manifest) []string {
	var problems []string
	problems = append(problems, validateHeader(m)...)
	problems = append(problems, validateLoopFileInventory(root, m.LoopFiles)...)
	seen := map[string]bool{}
	if len(m.Loops) == 0 {
		problems = append(problems, "loops or loop_files must provide at least one loop")
	}
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
	if strings.ContainsAny(path, "*?[") {
		matches, err := filepath.Glob(resolvePath(root, path))
		if err != nil {
			return []string{fmt.Sprintf("%s references invalid artifact glob %q: %v", owner, path, err)}
		}
		if len(matches) == 0 {
			return []string{fmt.Sprintf("%s references missing artifact glob %q", owner, path)}
		}
		return nil
	}
	if _, err := os.Stat(resolvePath(root, path)); err != nil {
		return []string{fmt.Sprintf("%s references missing artifact %q", owner, path)}
	}
	return nil
}
