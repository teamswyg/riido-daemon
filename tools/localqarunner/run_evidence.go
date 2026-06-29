package main

import "time"

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
		Artifacts:      newRunArtifacts(cfg),
	}
}

func newRunArtifacts(cfg config) runArtifacts {
	return runArtifacts{
		ProviderEvidence:  *cfg.providerEvidence,
		ProductEvidence:   *cfg.productEvidence,
		ReleaseEvidence:   *cfg.releaseEvidence,
		CoverageEvidence:  *cfg.coverageEvidence,
		PromotionRegistry: *cfg.promotionManifest,
		ManualEvidence:    *cfg.manualEvidence,
		DomainCache:       *cfg.domainCache,
		ProductLab:        *cfg.productLab,
		ScheduleEvidence:  *cfg.scheduleEvidence,
		InfraEvidence:     *cfg.infraEvidence,
		DashboardHTML:     *cfg.dashboardHTML,
		S3Prefix:          *cfg.s3Prefix,
	}
}
