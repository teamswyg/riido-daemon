package main

import (
	"fmt"
	"path/filepath"
	"time"
)

func run(cfg config) (string, error) {
	root, err := filepath.Abs(*cfg.repo)
	if err != nil {
		return statusFailed, fmt.Errorf("resolve repo root: %w", err)
	}
	start := time.Now().UTC()
	evidence := newEvidence(cfg, start)
	providerStatus := runProviderStep(root, cfg, &evidence)
	if providerStatus == statusFailed && !*cfg.continueOnFailure {
		return finishRun(root, cfg, evidence)
	}
	if *cfg.runRelease {
		releaseStatus := runReleaseStep(root, cfg, &evidence)
		if releaseStatus == statusFailed && !*cfg.continueOnFailure {
			return finishRun(root, cfg, evidence)
		}
	}
	if *cfg.runProduct {
		productStatus := runProductStep(root, cfg, &evidence)
		if productStatus == statusFailed && !*cfg.continueOnFailure {
			return finishRun(root, cfg, evidence)
		}
	}
	runScheduleInspectStep(root, cfg, &evidence)
	if _, err := finishRun(root, cfg, evidence); err != nil {
		return statusFailed, err
	}
	runDashboardStep(root, cfg, &evidence)
	if err := applyCoverageEvidence(root, cfg, &evidence); err != nil {
		return statusFailed, err
	}
	if *cfg.s3Prefix != "" {
		return runS3Phase(root, cfg, evidence)
	}
	if _, err := finishRun(root, cfg, evidence); err != nil {
		return statusFailed, err
	}
	runFinalDashboardStep(root, cfg, &evidence)
	if err := applyCoverageEvidence(root, cfg, &evidence); err != nil {
		return statusFailed, err
	}
	return finishRun(root, cfg, evidence)
}

func newEvidence(cfg config, observed time.Time) runEvidence {
	expires := observed.Add(*cfg.validFor)
	return runEvidence{
		SchemaVersion:  "riido-local-qa-run-result.v1",
		ID:             "local-qa-run",
		ObservedAt:     observed.Format(time.RFC3339),
		ExpiresAt:      expires.Format(time.RFC3339),
		Status:         statusPassed,
		CoverageStatus: statusPassed,
		StrictCoverage: boolValue(cfg.strictCoverage),
		Artifacts: runArtifacts{
			ProviderEvidence: *cfg.providerEvidence,
			ProductEvidence:  *cfg.productEvidence,
			ReleaseEvidence:  *cfg.releaseEvidence,
			CoverageEvidence: *cfg.coverageEvidence,
			ManualEvidence:   *cfg.manualEvidence,
			DomainCache:      *cfg.domainCache,
			ProductLab:       *cfg.productLab,
			ScheduleEvidence: *cfg.scheduleEvidence,
			InfraEvidence:    *cfg.infraEvidence,
			DashboardHTML:    *cfg.dashboardHTML,
			S3Prefix:         *cfg.s3Prefix,
		},
	}
}
