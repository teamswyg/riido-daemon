package project

func StateFromProjection(projection WorkspaceProjection) StateFile {
	state := baseStateFromProjection(projection)
	appendProjectStates(&state, projection.Projects)
	appendTaskStates(&state, projection.DocumentTaskLinks)
	ensureStateSlices(&state)
	return state
}

func baseStateFromProjection(projection WorkspaceProjection) StateFile {
	return StateFile{
		SchemaVersion:          StateSchemaVersion,
		ProjectionVersion:      projection.SchemaVersion,
		Root:                   projection.Root,
		Domain:                 projection.Domain,
		HarnessRunCount:        projection.HarnessRunCount,
		HarnessNextDirection:   projection.HarnessNextDirection,
		OrchestrationMode:      projection.OrchestrationMode,
		DecisionGate:           projection.DecisionGate,
		DecisionBy:             append([]string(nil), projection.DecisionBy...),
		DecisionLLMs:           append([]string(nil), projection.DecisionLLMs...),
		ProviderCandidates:     append([]ProviderCandidate(nil), projection.ProviderCandidates...),
		RecommendedProvider:    projection.RecommendedProvider,
		RecommendedDecisionLLM: projection.RecommendedDecisionLLM,
		NextAction:             projection.NextAction,
		Diagnostics:            append([]ProjectionDiagnostic(nil), projection.Diagnostics...),
	}
}

func ensureStateSlices(state *StateFile) {
	if state.Diagnostics == nil {
		state.Diagnostics = []ProjectionDiagnostic{}
	}
	if state.Projects == nil {
		state.Projects = []ProjectState{}
	}
	if state.Tasks == nil {
		state.Tasks = []TaskState{}
	}
}
