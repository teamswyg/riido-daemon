package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

const defaultManifest = "docs/30-architecture/package-workflow-evidence.riido.json"

func run(args []string) error {
	fs := flag.NewFlagSet("packageworkflowevidence", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	manifestPath := fs.String("manifest", defaultManifest, "package workflow evidence manifest")
	workflow := fs.String("workflow", "", "workflow path")
	evidenceOut := fs.String("evidence-out", "", "evidence output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *workflow == "" || *evidenceOut == "" {
		return errors.New("-workflow and -evidence-out are required")
	}
	m, err := loadManifest(*manifestPath)
	if err != nil {
		return err
	}
	spec, err := findWorkflow(m, *workflow)
	if err != nil {
		return err
	}
	body, err := os.ReadFile(*workflow)
	if err != nil {
		return fmt.Errorf("read workflow: %w", err)
	}
	value := buildEvidence(m, spec, string(body))
	if err := writeJSON(*evidenceOut, value); err != nil {
		return err
	}
	if value.Status != "verified" {
		return fmt.Errorf("%s evidence status %s", spec.ID, value.Status)
	}
	return nil
}
