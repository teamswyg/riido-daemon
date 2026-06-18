package project

import (
	"fmt"
	"sort"

	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
)

func appendProjectsFromMwsd(projection *WorkspaceProjection, repositories []mwsdbridge.ProjectRepository) {
	for _, repository := range repositories {
		project := projectFromRepository(repository)
		projection.Projects = append(projection.Projects, project)
		appendProjectDiagnostics(projection, project)
	}
	sort.Slice(projection.Projects, func(i, j int) bool {
		return projection.Projects[i].ID < projection.Projects[j].ID
	})
}

func projectFromRepository(repository mwsdbridge.ProjectRepository) Project {
	return Project{
		ID:            repository.Name,
		Owner:         repository.Owner,
		Visibility:    repository.Visibility,
		SSOTScope:     repository.SSOTScope,
		LocalPath:     repository.LocalPath,
		Remote:        repository.Remote,
		Role:          repository.Role,
		Consumes:      append([]string(nil), repository.Consumes...),
		Health:        repositoryHealth(repository),
		LocalPresent:  repository.LocalPresent,
		GitPresent:    repository.GitPresent,
		RemoteMatches: repository.RemoteMatches,
	}
}

func appendProjectDiagnostics(projection *WorkspaceProjection, project Project) {
	if project.Visibility != "private" {
		projection.Diagnostics = append(projection.Diagnostics, ProjectionDiagnostic{
			Severity: "warning",
			Code:     "project-not-private",
			Message:  fmt.Sprintf("project %s visibility is %s", project.ID, project.Visibility),
		})
	}
	if project.Health != RepositoryReady {
		projection.Diagnostics = append(projection.Diagnostics, ProjectionDiagnostic{
			Severity: "error",
			Code:     "project-repository-not-ready",
			Message:  fmt.Sprintf("project %s repository health is %s", project.ID, project.Health),
		})
	}
}
