package main

import "github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"

func countRuntimeModelDefaults(models []runtimeactor.RuntimeModel) int {
	count := 0
	for _, model := range models {
		if model.IsDefault {
			count++
		}
	}
	return count
}

func runtimeDefaultModelID(models []runtimeactor.RuntimeModel) string {
	for _, model := range models {
		if model.IsDefault {
			return model.ModelID
		}
	}
	return ""
}
