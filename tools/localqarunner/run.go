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
	runDashboardStep(root, cfg, &evidence)
	if *cfg.s3Prefix != "" {
		if _, err := finishRun(root, cfg, evidence); err != nil {
			return statusFailed, err
		}
		runUploadSteps(root, cfg, &evidence)
	}
	return finishRun(root, cfg, evidence)
}

func newEvidence(cfg config, observed time.Time) runEvidence {
	expires := observed.Add(*cfg.validFor)
	return runEvidence{
		SchemaVersion: "riido-local-qa-run-result.v1",
		ID:            "local-qa-run",
		ObservedAt:    observed.Format(time.RFC3339),
		ExpiresAt:     expires.Format(time.RFC3339),
		Status:        statusPassed,
		Artifacts: runArtifacts{
			ProviderEvidence: *cfg.providerEvidence,
			ProductEvidence:  *cfg.productEvidence,
			ReleaseEvidence:  *cfg.releaseEvidence,
			ProductLab:       *cfg.productLab,
			ScheduleEvidence: *cfg.scheduleEvidence,
			DashboardHTML:    *cfg.dashboardHTML,
			S3Prefix:         *cfg.s3Prefix,
		},
	}
}
