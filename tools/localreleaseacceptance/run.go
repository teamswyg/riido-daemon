package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

func run(ctx context.Context, opts options) error {
	root, err := filepath.Abs(opts.Repo)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}
	fixture, cleanup, err := newInstallFixture()
	if err != nil {
		return err
	}
	defer cleanup()
	observed := time.Now().UTC()
	result := newEvidence(observed, opts.ValidFor)
	result.Artifacts.InstallDir = fixture.installDir
	result.Artifacts.InstalledBinary = filepath.Join(fixture.installDir, "riido")
	scenario, version := verifyInstaller(ctx, root, fixture)
	result.Artifacts.VersionOutput = version
	result.Scenarios = append(result.Scenarios, scenario)
	result.Status = scenario.Status
	if opts.EvidenceOut != "" {
		if err := writeJSON(outputPath(root, opts.EvidenceOut), result); err != nil {
			return err
		}
	}
	if scenario.Status != statusPassed {
		return fmt.Errorf("release acceptance failed: %s", scenario.FailureSummary)
	}
	return nil
}

func newEvidence(observed time.Time, validFor time.Duration) evidenceFile {
	return evidenceFile{
		SchemaVersion: "riido-local-release-acceptance.v1",
		ID:            "local-release-acceptance",
		ObservedAt:    observed.Format(time.RFC3339),
		ExpiresAt:     observed.Add(validFor).Format(time.RFC3339),
		Status:        statusPassed,
	}
}
