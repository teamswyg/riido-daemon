package main

import (
	"fmt"
	"os"
)

func verifyResult(result auditResult) error {
	if len(result.Unregistered) > 0 {
		return fmt.Errorf("unregistered workflow evidence gaps: %v", result.Unregistered)
	}
	if len(result.NonStrict) > 0 {
		return fmt.Errorf("non-strict artifact upload steps: %v", result.NonStrict)
	}
	if len(result.MissingEvidence) > 0 {
		return fmt.Errorf("missing evidence artifact uploads: %v", result.MissingEvidence)
	}
	if len(result.MissingEvidenceTools) > 0 {
		return fmt.Errorf("evidence tools missing workflow coverage: %v", result.MissingEvidenceTools)
	}
	if len(result.MissingEvidenceToolBindings) > 0 {
		return fmt.Errorf("evidence tools missing evidence-out artifact binding: %v",
			result.MissingEvidenceToolBindings)
	}
	if len(result.AcceptedUnused) > 0 {
		return fmt.Errorf("unused accepted workflow gaps: %v", result.AcceptedUnused)
	}
	return nil
}

func verifyDoc(root string, m manifest, expected string) error {
	current, err := os.ReadFile(repoPath(root, m.GeneratedDoc))
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(current) != expected {
		return fmt.Errorf("generated doc drift: run go run ./tools/workflowevidence -write-doc")
	}
	return nil
}
