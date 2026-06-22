package main

import "path/filepath"

func runProviderStep(root string, cfg config, evidence *runEvidence) string {
	args := []string{"run", *cfg.providerTool, "-repo", root, "-check-doc"}
	if *cfg.runIntegration {
		args = append(args, "-run-integration")
	}
	args = append(args,
		"-valid-for", cfg.validFor.String(),
		"-evidence-out", *cfg.providerEvidence,
	)
	step := runStep(root, "provider-evidence", "go", args...)
	appendStep(evidence, step)
	return step.Status
}

func runDashboardStep(root string, cfg config, evidence *runEvidence) {
	args := []string{
		"run", *cfg.dashboardTool,
		"-provider-evidence", *cfg.providerEvidence,
		"-coverage-manifest", *cfg.coverageManifest,
		"-out", *cfg.dashboardHTML,
	}
	if *cfg.productEvidence != "" {
		args = append(args, "-product-evidence", *cfg.productEvidence)
	}
	appendStep(evidence, runStep(root, "dashboard-render", "go", args...))
}

func runUploadSteps(root string, cfg config, evidence *runEvidence) {
	stamp := timestampSlug(evidence.ObservedAt)
	for _, upload := range uploads(cfg, stamp) {
		args := []string{"s3", "cp", upload.source, upload.target}
		args = append(args, "--cache-control", "no-store")
		appendStep(evidence, runStep(root, upload.id, "aws", args...))
	}
}

func uploads(cfg config, stamp string) []uploadSpec {
	prefix := trimTrailingSlash(*cfg.s3Prefix)
	specs := []uploadSpec{
		upload("dashboard-html", *cfg.dashboardHTML, prefix+"/latest/index.html"),
		upload("provider-evidence", *cfg.providerEvidence, prefix+"/latest/provider-real-cli-observation.json"),
		upload("run-evidence", *cfg.runEvidence, prefix+"/latest/local-qa-run.json"),
		upload("dashboard-html-"+stamp, *cfg.dashboardHTML, prefix+"/"+stamp+"/index.html"),
		upload("provider-evidence-"+stamp, *cfg.providerEvidence, prefix+"/"+stamp+"/provider-real-cli-observation.json"),
		upload("run-evidence-"+stamp, *cfg.runEvidence, prefix+"/"+stamp+"/local-qa-run.json"),
	}
	if *cfg.productEvidence != "" {
		specs = append(specs,
			upload("product-evidence", *cfg.productEvidence, prefix+"/latest/ai-agent-product-acceptance.json"),
			upload("product-evidence-"+stamp, *cfg.productEvidence, prefix+"/"+stamp+"/ai-agent-product-acceptance.json"),
		)
	}
	return specs
}

func upload(id, source, target string) uploadSpec {
	return uploadSpec{id: "upload-" + id, source: source, target: target}
}

func runEvidenceAbs(root string, cfg config) string {
	return outputPath(root, *cfg.runEvidence)
}

func outputPath(root, path string) string {
	clean := filepath.FromSlash(path)
	if filepath.IsAbs(clean) {
		return clean
	}
	return filepath.Join(root, clean)
}
