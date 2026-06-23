package main

func uploads(cfg config, stamp string) []uploadSpec {
	prefix := trimTrailingSlash(*cfg.s3Prefix)
	specs := baseUploads(cfg, stamp, prefix)
	if *cfg.productEvidence != "" && fileExists(*cfg.productEvidence) {
		specs = append(specs,
			upload("product-evidence", *cfg.productEvidence, prefix+"/latest/ai-agent-product-acceptance.json"),
			upload("product-evidence-"+stamp, *cfg.productEvidence, prefix+"/"+stamp+"/ai-agent-product-acceptance.json"),
		)
	}
	if *cfg.releaseEvidence != "" && fileExists(*cfg.releaseEvidence) {
		specs = append(specs,
			upload("release-evidence", *cfg.releaseEvidence, prefix+"/latest/local-release-acceptance.json"),
			upload("release-evidence-"+stamp, *cfg.releaseEvidence, prefix+"/"+stamp+"/local-release-acceptance.json"),
		)
	}
	if *cfg.coverageEvidence != "" && fileExists(*cfg.coverageEvidence) {
		specs = append(specs,
			upload("coverage-evidence", *cfg.coverageEvidence, prefix+"/latest/local-qa-coverage.json"),
			upload("coverage-evidence-"+stamp, *cfg.coverageEvidence, prefix+"/"+stamp+"/local-qa-coverage.json"),
		)
	}
	if *cfg.manualEvidence != "" && fileExists(*cfg.manualEvidence) {
		specs = append(specs,
			upload("manual-evidence", *cfg.manualEvidence, prefix+"/latest/manual-qa-evidence.json"),
			upload("manual-evidence-"+stamp, *cfg.manualEvidence, prefix+"/"+stamp+"/manual-qa-evidence.json"),
		)
	}
	if *cfg.domainCache != "" && fileExists(*cfg.domainCache) {
		specs = append(specs,
			upload("domain-cache", *cfg.domainCache, prefix+"/latest/domain-fixture-journey-cache.json"),
			upload("domain-cache-"+stamp, *cfg.domainCache, prefix+"/"+stamp+"/domain-fixture-journey-cache.json"),
		)
	}
	if *cfg.productLab != "" && fileExists(*cfg.productLab) {
		specs = append(specs,
			upload("product-lab", *cfg.productLab, prefix+"/latest/contract-lab/index.html"),
			upload("product-lab-"+stamp, *cfg.productLab, prefix+"/"+stamp+"/contract-lab/index.html"),
		)
	}
	if *cfg.scheduleEvidence != "" && fileExists(*cfg.scheduleEvidence) {
		specs = append(specs,
			upload("schedule-evidence", *cfg.scheduleEvidence, prefix+"/latest/local-qa-schedule.json"),
			upload("schedule-evidence-"+stamp, *cfg.scheduleEvidence, prefix+"/"+stamp+"/local-qa-schedule.json"),
		)
	}
	if *cfg.infraEvidence != "" && fileExists(*cfg.infraEvidence) {
		specs = append(specs,
			upload("infra-evidence", *cfg.infraEvidence, prefix+"/latest/local-qa-dashboard-infra-evidence.json"),
			upload("infra-evidence-"+stamp, *cfg.infraEvidence, prefix+"/"+stamp+"/local-qa-dashboard-infra-evidence.json"),
		)
	}
	if *cfg.productScreenshots != "" && dirExists(*cfg.productScreenshots) {
		specs = append(specs,
			uploadDir("product-screenshots", *cfg.productScreenshots, prefix+"/latest/screenshots/ai-agent-product-acceptance/"),
			uploadDir("product-screenshots-"+stamp, *cfg.productScreenshots, prefix+"/"+stamp+"/screenshots/ai-agent-product-acceptance/"),
		)
	}
	return specs
}

func baseUploads(cfg config, stamp, prefix string) []uploadSpec {
	return []uploadSpec{
		upload("dashboard-html", *cfg.dashboardHTML, prefix+"/latest/index.html"),
		upload("provider-evidence", *cfg.providerEvidence, prefix+"/latest/provider-real-cli-observation.json"),
		upload("run-evidence", *cfg.runEvidence, prefix+"/latest/local-qa-run.json"),
		upload("dashboard-html-"+stamp, *cfg.dashboardHTML, prefix+"/"+stamp+"/index.html"),
		upload("provider-evidence-"+stamp, *cfg.providerEvidence, prefix+"/"+stamp+"/provider-real-cli-observation.json"),
		upload("run-evidence-"+stamp, *cfg.runEvidence, prefix+"/"+stamp+"/local-qa-run.json"),
	}
}

func upload(id, source, target string) uploadSpec {
	return uploadSpec{id: "upload-" + id, source: source, target: target}
}

func uploadDir(id, source, target string) uploadSpec {
	return uploadSpec{id: "upload-" + id, source: source, target: target, recursive: true}
}
