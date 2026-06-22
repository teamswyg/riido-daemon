package main

func cleanupPartialTaskAssignments(
	client apiClient,
	base string,
	taskID string,
	run taskAssignmentRun,
) []scenario {
	if run.First.ID == "" {
		return nil
	}
	return []scenario{
		stopAssignmentWithID(client, base, taskID, run.Pair.First.AgentID,
			"contract.task.assignment.cleanup.partial.stop"),
		deleteAssignment(client, base, taskID, run.Pair.First.AgentID,
			"contract.task.assignment.cleanup.partial.delete"),
	}
}
