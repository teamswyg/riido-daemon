package main

func runScheduleInspectStep(root string, cfg config, evidence *runEvidence) {
	if !fileExists(*cfg.scheduleEvidence) {
		return
	}
	args := []string{
		"run", *cfg.scheduleTool,
		"-repo", root,
		"-inspect",
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
	args = appendScheduleProductArgs(args, cfg)
	appendStep(evidence, runStep(root, "schedule-inspect", "go", args...))
}

func appendScheduleProductArgs(args []string, cfg config) []string {
	if !*cfg.runProduct {
		args = append(args, "-run-product=false")
	}
	if *cfg.productStartClient {
		args = append(args, "-product-start-client")
	}
	if *cfg.productWorkspace != "" {
		args = append(args, "-product-workspace-id", *cfg.productWorkspace)
	}
	if *cfg.productTeamID != "" {
		args = append(args, "-product-team-id", *cfg.productTeamID)
	}
	return appendScheduleTaskArgs(args, cfg)
}

func appendScheduleTaskArgs(args []string, cfg config) []string {
	if *cfg.productTaskID != "" {
		args = append(args, "-product-task-id", *cfg.productTaskID)
	}
	if *cfg.productAgentID1 != "" {
		args = append(args, "-product-agent-id-1", *cfg.productAgentID1)
	}
	if *cfg.productAgentID2 != "" {
		args = append(args, "-product-agent-id-2", *cfg.productAgentID2)
	}
	if *cfg.productCommentBody != "" {
		args = append(args, "-product-comment-body", *cfg.productCommentBody)
	}
	if !*cfg.productMutations {
		args = append(args, "-product-task-mutations=false")
	}
	if !*cfg.productTaskFixture {
		args = append(args, "-product-create-task-fixture=false")
	}
	return args
}
