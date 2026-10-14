package main

import "path/filepath"

func scheduleStepArgs(root string, cfg config) ([]string, string) {
	args := []string{
		"run", *cfg.scheduleTool,
		"-repo", root,
		"-evidence-out", *cfg.scheduleEvidence,
		"-s3-prefix", *cfg.s3Prefix,
		"-client-root", *cfg.clientRoot,
		"-product-base-url", *cfg.productBaseURL,
		"-product-agent-host", *cfg.productAgentHost,
		"-product-riido-api-host", *cfg.productRiidoHost,
		"-product-storage-state", *cfg.productStorage,
		"-product-evidence", *cfg.productEvidence,
		"-coverage-evidence", *cfg.coverageEvidence,
	}
	id := "schedule-evidence"
	if fileExists(*cfg.scheduleEvidence) {
		id = "schedule-inspect"
		args = append(args, "-inspect")
	} else {
		args = append(args, "-plist", filepath.Join(root, ".riido-local", "local-qa.plist"))
	}
	return appendScheduleProductArgs(args, cfg), id
}
