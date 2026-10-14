package project

func appendTaskStates(state *StateFile, links []DocumentTaskLink) {
	for _, link := range links {
		state.Tasks = append(state.Tasks, taskStateFromDocumentLink(link))
	}
}

func taskStateFromDocumentLink(link DocumentTaskLink) TaskState {
	return TaskState{
		ID:                     link.TaskID,
		ProjectID:              link.ProjectID,
		State:                  "Created",
		SourceDocumentID:       link.DocumentID,
		SourceDocumentPath:     link.DocumentPath,
		Title:                  link.Title,
		Owner:                  link.Owner,
		SourceStatus:           link.Status,
		RecommendedProvider:    link.RecommendedProvider,
		RecommendedDecisionLLM: link.RecommendedDecisionLLM,
		RequiresHumanApproval:  link.RequiresHumanApproval,
		HarnessNextDirection:   link.HarnessNextDirection,
	}
}
