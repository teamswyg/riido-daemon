package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyWorkflow(root string, m manifest) error {
	body, err := os.ReadFile(repoPath(root, m.Workflow))
	if err != nil {
		return err
	}
	for _, needle := range []string{
		"go run ./tools/localqacandidatedecision",
		"-candidate-in out/local-qa-run.json",
		"-github-annotations",
		m.EvidenceArtifact,
	} {
		if !strings.Contains(string(body), needle) {
			return fmt.Errorf("local QA candidate decision workflow missing %q", needle)
		}
	}
	return nil
}
