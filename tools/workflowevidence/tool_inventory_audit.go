package main

func auditEvidenceTools(root string, workflows []workflowSource) (int, int, int, []string, []string) {
	tools := evidenceToolDirs(root)
	called, bound := 0, 0
	texts := workflowSourceTexts(workflows)
	var missingCalls, missingBindings []string
	for _, tool := range tools {
		if workflowCallsEvidenceTool(texts, tool) {
			called++
		} else {
			missingCalls = append(missingCalls, tool)
		}
		if workflowBindsEvidenceTool(workflows, tool) {
			bound++
		} else {
			missingBindings = append(missingBindings, tool)
		}
	}
	return len(tools), called, bound, uniqueStrings(missingCalls), uniqueStrings(missingBindings)
}

func workflowSourceTexts(workflows []workflowSource) []string {
	texts := make([]string, 0, len(workflows))
	for _, workflow := range workflows {
		texts = append(texts, workflow.Text)
	}
	return texts
}
