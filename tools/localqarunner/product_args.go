package main

func appendProductTaskArgs(args []string, cfg config) []string {
	args = append(args, "-task-id", *cfg.productTaskID)
	args = append(args, "-first-agent-id", *cfg.productAgentID1)
	args = append(args, "-second-agent-id", *cfg.productAgentID2)
	args = append(args, "-comment-body", *cfg.productCommentBody)
	if *cfg.productMutations {
		args = append(args, "-run-task-mutations")
	}
	return args
}
