package project

func appendProjectStates(state *StateFile, projects []Project) {
	for _, project := range projects {
		state.Projects = append(state.Projects, projectStateFromProjection(project))
	}
}

func projectStateFromProjection(project Project) ProjectState {
	return ProjectState{
		ID:            project.ID,
		Owner:         project.Owner,
		Visibility:    project.Visibility,
		SSOTScope:     project.SSOTScope,
		LocalPath:     project.LocalPath,
		Remote:        project.Remote,
		Role:          project.Role,
		Health:        project.Health,
		LocalPresent:  project.LocalPresent,
		GitPresent:    project.GitPresent,
		RemoteMatches: project.RemoteMatches,
	}
}
