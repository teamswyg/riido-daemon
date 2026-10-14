package main

import (
	"fmt"
	"slices"
)

func verifyAll(root string, m manifest) (verifyResult, error) {
	if err := verifyIdentity(m); err != nil {
		return verifyResult{}, err
	}
	if err := verifyLoop(m.Loop); err != nil {
		return verifyResult{}, err
	}
	if err := verifyWorkflow(root, m); err != nil {
		return verifyResult{}, err
	}
	if err := verifyDecisions(m); err != nil {
		return verifyResult{}, err
	}
	return verifyResult{
		DecisionCount:         len(m.Decisions),
		ManifestDecisionCount: len(m.Decisions),
	}, nil
}

func verifyIdentity(m manifest) error {
	if m.SchemaVersion != manifestSchema || m.ID != requiredID {
		return fmt.Errorf("unexpected local QA candidate decision identity")
	}
	if m.EvidenceTool != "tools/localqacandidatedecision" {
		return fmt.Errorf("evidence_tool must be tools/localqacandidatedecision")
	}
	if m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		return fmt.Errorf("generated_doc, workflow, and evidence_artifact are required")
	}
	return nil
}

func verifyLoop(loop evidenceLoop) error {
	if slices.Contains([]string{
		loop.Observation, loop.Hypothesis, loop.Execute,
		loop.Evaluate, loop.Retrospective,
	}, "") {
		return fmt.Errorf("loop evidence fields are required")
	}
	return nil
}
