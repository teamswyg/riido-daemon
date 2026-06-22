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
