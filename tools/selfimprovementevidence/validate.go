package main

import (
	"errors"
	"fmt"
)

func validateManifest(m manifest) error {
	if m.SchemaVersion != manifestSchema {
		return fmt.Errorf("schema_version = %q, want %q", m.SchemaVersion, manifestSchema)
	}
	if m.ID == "" || m.Title == "" || m.GeneratedDoc == "" {
		return errors.New("id, title, and generated_doc are required")
	}
	if m.Workflow == "" || m.EvidenceArtifact == "" || m.LoopSource == "" {
		return errors.New("workflow, evidence_artifact, and loop_source are required")
	}
	if len(m.Required) == 0 {
		return errors.New("required_evidence is required")
	}
	seen := map[string]bool{}
	for _, item := range m.Required {
		if err := validateRequired(item, seen); err != nil {
			return err
		}
	}
	return nil
}

func validateRequired(item requiredEvidence, seen map[string]bool) error {
	if item.ID == "" || item.File == "" || item.Description == "" {
		return errors.New("required evidence id, file, and description are required")
	}
	if seen[item.ID] {
		return fmt.Errorf("duplicate required evidence %s", item.ID)
	}
	seen[item.ID] = true
	if len(item.Assertions) == 0 {
		return fmt.Errorf("%s has no assertions", item.ID)
	}
	for _, check := range item.Assertions {
		if check.Field == "" {
			return fmt.Errorf("%s has assertion without field", item.ID)
		}
	}
	return nil
}
