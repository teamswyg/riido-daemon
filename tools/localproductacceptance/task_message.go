package main

import "net/http"

func threadMessageScenario(client apiClient, base string, plan taskMutationPlan, assigned scenario) scenario {
	threadID, _ := assigned.Observed["thread_id"].(string)
	if threadID == "" {
		return failTaskScenario("contract.task.thread_message", "assignment did not return a thread_id")
	}
	path := taskEndpoint(base, plan.TaskID, "/threads/"+threadID+"/messages")
	body := map[string]any{"body": plan.CommentBody}
	return apiQuery(client, "contract.task.thread_message", http.MethodPost, path, body, summarizeTaskAction)
}
