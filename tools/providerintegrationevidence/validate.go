package main

import (
	"fmt"
	"os"
)

func validateManifest(root string, m manifest) []string {
	var problems []string
	if m.SchemaVersion != "riido-provider-real-cli-observation.v1" {
		problems = append(problems, "schema_version must be riido-provider-real-cli-observation.v1")
	}
	for _, value := range []string{m.ID, m.Title, m.GeneratedDoc, m.Workflow, m.EvidenceArtifact} {
		if value == "" {
			problems = append(problems, "id, title, generated_doc, workflow, and evidence_artifact are required")
		}
	}
	problems = append(problems, mustExist(root, m.Workflow)...)
	if len(m.Providers) == 0 {
		problems = append(problems, "providers must not be empty")
	}
	seen := map[string]bool{}
	for _, provider := range m.Providers {
		problems = append(problems, validateProvider(provider, seen)...)
		problems = append(problems, mustExist(root, provider.GoPackage)...)
	}
	return problems
}

func validateProvider(provider provider, seen map[string]bool) []string {
	var problems []string
	if provider.ID == "" || provider.DisplayName == "" || provider.DefaultExecutable == "" ||
		provider.OverrideEnv == "" || provider.GoPackage == "" || provider.TestRegex == "" {
		problems = append(problems, "provider rows require id, display_name, executable, env, package, and test")
	}
	if seen[provider.ID] {
		problems = append(problems, fmt.Sprintf("duplicate provider id %q", provider.ID))
	}
	seen[provider.ID] = true
	return problems
}

func mustExist(root, path string) []string {
	if path == "" {
		return nil
	}
	if _, err := os.Stat(resolvePath(root, path)); err != nil {
		return []string{fmt.Sprintf("missing artifact %q", path)}
	}
	return nil
}
