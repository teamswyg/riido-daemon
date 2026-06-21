package main

import (
	"fmt"
	"os"
)

const (
	defaultManifest = "docs/30-architecture/ci-evidence.riido.json"
	evidenceSchema  = "riido-daemon-ci-evidence.v1"
)

func run(opts options) error {
	if opts.Workflow == "" || opts.ID == "" || opts.EvidenceOut == "" {
		return fmt.Errorf("workflow, id, and evidence-out are required")
	}
	m, err := loadManifest(opts.Manifest)
	if err != nil {
		return err
	}
	spec, err := findWorkflow(m, opts.Workflow, opts.ID)
	if err != nil {
		return err
	}
	body, err := os.ReadFile(opts.Workflow)
	if err != nil {
		return fmt.Errorf("read workflow: %w", err)
	}
	report := buildEvidence(m, spec, string(body))
	if err := writeJSON(opts.EvidenceOut, report); err != nil {
		return err
	}
	if report.Status != "verified" {
		return fmt.Errorf("ci evidence invalid: %v", report.Problems)
	}
	return nil
}

func buildEvidence(m manifest, spec workflowSpec, text string) evidence {
	report := evidence{
		SchemaVersion:    evidenceSchema,
		ID:               spec.ID,
		Status:           "verified",
		Workflow:         spec.Workflow,
		EvidenceArtifact: spec.EvidenceArtifact,
		LoopSource:       m.LoopSource,
		Problems:         []string{},
	}
	for _, command := range spec.RequiredCommands {
		found := workflowContainsCommand(text, command)
		report.Required = append(report.Required, required{Command: command, Found: found})
		if !found {
			report.Problems = append(report.Problems, "missing command "+command)
		}
	}
	report.ProblemCount = len(report.Problems)
	if len(report.Problems) > 0 {
		report.Status = "failed"
	}
	return report
}
