package main

func runReleaseStep(root string, cfg config, evidence *runEvidence) string {
	args := []string{
		"run", *cfg.releaseTool,
		"-repo", root,
		"-valid-for", cfg.validFor.String(),
		"-evidence-out", *cfg.releaseEvidence,
	}
	step := runStep(root, "release-acceptance", "go", args...)
	appendStep(evidence, step)
	return step.Status
}
