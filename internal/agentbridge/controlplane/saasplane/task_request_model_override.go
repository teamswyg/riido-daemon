package saasplane

import (
	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
)

func assignmentProviderModelOverride(assignment assignmentcontract.Assignment) string {
	if usesLocalCodexModelConfig(assignment.RuntimeProvider) {
		return ""
	}
	return providercatalog.ModelOverride(assignment.RuntimeProvider, assignment.ModelID)
}

func usesLocalCodexModelConfig(provider string) bool {
	return provider == "codex"
}
