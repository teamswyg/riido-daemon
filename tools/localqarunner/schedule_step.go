package main

var runLocalQAScheduleStep = runStep

func runScheduleInspectStep(root string, cfg config, evidence *runEvidence) {
	args, id := scheduleStepArgs(root, cfg)
	appendStep(evidence, runLocalQAScheduleStep(root, id, "go", args...))
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
