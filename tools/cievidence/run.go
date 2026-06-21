package main

import (
	"fmt"
	"os"
)

func run(opts options) error {
	if opts.Workflow == "" || opts.ID == "" || opts.EvidenceOut == "" {
		return fmt.Errorf("workflow, id, and evidence-out are required")
	}
	body, err := os.ReadFile(opts.Workflow)
	if err != nil {
		return fmt.Errorf("read workflow: %w", err)
	}
	report := buildEvidence(opts, string(body))
	if err := writeJSON(opts.EvidenceOut, report); err != nil {
		return err
	}
	if report.Status != "verified" {
		return fmt.Errorf("ci evidence invalid: %v", report.Problems)
	}
	return nil
}

func buildEvidence(opts options, text string) evidence {
	requiredCommands := requiredCommandsFor(opts.Workflow)
	report := evidence{
		SchemaVersion: "riido-daemon-ci-evidence.v1",
		ID:            opts.ID,
		Status:        "verified",
		Workflow:      opts.Workflow,
	}
	for _, command := range requiredCommands {
		found := workflowContainsCommand(text, command)
		report.Required = append(report.Required, required{Command: command, Found: found})
		if !found {
			report.Problems = append(report.Problems, "missing command "+command)
		}
	}
	if len(report.Problems) > 0 {
		report.Status = "failed"
	}
	return report
}
