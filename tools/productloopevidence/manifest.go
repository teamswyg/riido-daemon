package main

import "fmt"

func loadManifest(path string) (manifest, error) {
	var m manifest
	if err := loadJSON(path, &m); err != nil {
		return manifest{}, err
	}
	return m, validateManifest(m)
}

func validateManifest(m manifest) error {
	if m.SchemaVersion != manifestSchema {
		return fmt.Errorf("schema_version = %q, want %q", m.SchemaVersion, manifestSchema)
	}
	if m.ID == "" || m.Title == "" || m.GeneratedDoc == "" {
		return fmt.Errorf("id, title, and generated_doc are required")
	}
	if m.Workflow == "" || m.EvidenceArtifact == "" {
		return fmt.Errorf("workflow and evidence_artifact are required")
	}
	if m.LoopRegistry == "" || m.EntrypointRouteMap == "" {
		return fmt.Errorf("loop_registry and entrypoint_route_map are required")
	}
	if m.LocalAcceptanceManifest == "" || m.QASystemManifest == "" || m.LocalQAScheduleManifest == "" {
		return fmt.Errorf("local_acceptance_manifest, qa_system_manifest, and local_qa_schedule_manifest are required")
	}
	if m.Thresholds.MaxEntrypointsBeforePartial <= 0 || m.Thresholds.StalePartialAfterDays <= 0 {
		return fmt.Errorf("positive thresholds are required")
	}
	if len(m.OutcomeSignals) == 0 {
		return fmt.Errorf("outcome_signals are required")
	}
	return validateLoop(m.Loop)
}

func validateLoop(loop evidenceLoop) error {
	if loop.Observation == "" || loop.Hypothesis == "" || loop.Execute == "" {
		return fmt.Errorf("loop observation, hypothesis, and execute are required")
	}
	if loop.Evaluate == "" || loop.Retrospective == "" {
		return fmt.Errorf("loop evaluate and retrospective are required")
	}
	return nil
}
