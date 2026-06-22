package main

import "net/http"

func cleanupTaskAssignments(client apiClient, base string, plan taskMutationPlan) []scenario {
	return []scenario{
		stopAssignment(client, base, plan.TaskID, plan.Pair.First.AgentID),
		deleteAssignment(client, base, plan.TaskID, plan.Pair.First.AgentID, "contract.task.assignment.cleanup.first.delete"),
		deleteAssignment(client, base, plan.TaskID, plan.Pair.Second.AgentID, "contract.task.assignment.cleanup.second.delete"),
	}
}

func stopAssignment(client apiClient, base, taskID, agentID string) scenario {
	path := taskEndpoint(base, taskID, "/agent-assignments/"+agentID+"/stop")
	return apiQuery(client, "contract.task.assignment.cleanup.first.stop",
		http.MethodPost, path, nil, summarizeTaskAction)
}

func deleteAssignment(client apiClient, base, taskID, agentID, id string) scenario {
	path := taskEndpoint(base, taskID, "/agent-assignments/"+agentID)
	return apiQuery(client, id, http.MethodDelete, path, nil, summarizeTaskAction)
}
