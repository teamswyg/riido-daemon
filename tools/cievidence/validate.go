package main

import (
	"errors"
	"fmt"
)

const manifestSchema = "riido-daemon-ci-evidence.v1"

func validateManifest(m manifest) error {
	if m.SchemaVersion != manifestSchema {
		return fmt.Errorf("schema_version = %q, want %q", m.SchemaVersion, manifestSchema)
	}
	if m.ID == "" || m.LoopSource == "" || len(m.Workflows) == 0 {
		return errors.New("id, loop_source, and workflows are required")
	}
	seen := map[string]bool{}
	for _, spec := range m.Workflows {
		if err := validateWorkflowSpec(spec, seen); err != nil {
			return err
		}
	}
	return nil
}

func validateWorkflowSpec(spec workflowSpec, seen map[string]bool) error {
	if spec.ID == "" || spec.Workflow == "" || spec.EvidenceArtifact == "" {
		return errors.New("workflow id, workflow, and evidence_artifact are required")
	}
	if len(spec.RequiredCommands) == 0 {
		return fmt.Errorf("%s has no required commands", spec.ID)
	}
	if seen[spec.Workflow] {
		return fmt.Errorf("duplicate workflow %s", spec.Workflow)
	}
	seen[spec.Workflow] = true
	return nil
}
