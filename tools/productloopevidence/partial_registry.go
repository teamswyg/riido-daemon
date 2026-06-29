package main

func inferenceRequiredIDs(items []qaExecution) []string {
	var out []string
	for _, item := range items {
		if item.Mode != "" && item.Mode != "system" {
			out = append(out, item.ID)
		}
	}
	return out
}

func candidateLoopIDs(items []registryLoop) []string {
	var out []string
	for _, item := range items {
		if item.Kind != "closed-loop" {
			out = append(out, item.ID)
		}
	}
	return out
}

func promotedCount(items []registryLoop) int {
	count := 0
	for _, item := range items {
		if item.Kind == "closed-loop" {
			count++
		}
	}
	return count
}
