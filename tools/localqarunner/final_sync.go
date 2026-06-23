package main

import "fmt"

var runFinalSyncStep = runStep

func syncFinalDashboardArtifacts(root string, cfg config, evidence *runEvidence) error {
	prefix := trimTrailingSlash(*cfg.s3Prefix)
	stamp := timestampSlug(evidence.ObservedAt)
	for _, upload := range finalDashboardUploads(cfg, stamp, prefix) {
		args := []string{"s3", "cp", upload.source, upload.target, "--cache-control", "no-store"}
		step := runFinalSyncStep(root, upload.id, "aws", args...)
		appendStep(evidence, step)
		if step.Status != statusPassed {
			return fmt.Errorf("%s failed: %s", step.ID, step.OutputTail)
		}
	}
	return nil
}

func syncFinalRunEvidence(root string, cfg config, observedAt string) error {
	prefix := trimTrailingSlash(*cfg.s3Prefix)
	stamp := timestampSlug(observedAt)
	for _, upload := range finalRunEvidenceUploads(cfg, stamp, prefix) {
		args := []string{"s3", "cp", upload.source, upload.target, "--cache-control", "no-store"}
		step := runFinalSyncStep(root, upload.id, "aws", args...)
		if step.Status != statusPassed {
			return fmt.Errorf("%s failed: %s", step.ID, step.OutputTail)
		}
	}
	return nil
}

func finalDashboardUploads(cfg config, stamp, prefix string) []uploadSpec {
	return []uploadSpec{
		upload("dashboard-html-final", *cfg.dashboardHTML, prefix+"/latest/index.html"),
		upload("coverage-evidence-final", *cfg.coverageEvidence, prefix+"/latest/local-qa-coverage.json"),
		upload("dashboard-html-final-"+stamp, *cfg.dashboardHTML, prefix+"/"+stamp+"/index.html"),
		upload("coverage-evidence-final-"+stamp, *cfg.coverageEvidence, prefix+"/"+stamp+"/local-qa-coverage.json"),
	}
}

func finalRunEvidenceUploads(cfg config, stamp, prefix string) []uploadSpec {
	return []uploadSpec{
		upload("run-evidence-final", *cfg.runEvidence, prefix+"/latest/local-qa-run.json"),
		upload("run-evidence-final-"+stamp, *cfg.runEvidence, prefix+"/"+stamp+"/local-qa-run.json"),
	}
}
