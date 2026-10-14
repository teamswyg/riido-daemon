package project

import (
	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
)

func providerRuns(runs []mwsdbridge.OrchestrationRun) []ProviderRunSummary {
	out := make([]ProviderRunSummary, 0, len(runs))
	for _, run := range runs {
		out = append(out, ProviderRunSummary{
			ID:        run.ID,
			Direction: run.Direction,
			Source:    run.Source,
			Provider:  run.Provider,
			Command:   run.Command,
			Result:    run.Result,
		})
	}
	return out
}

func hasProject(projects []Project, id string) bool {
	for _, project := range projects {
		if project.ID == id {
			return true
		}
	}
	return false
}

func repositoryHealth(repository mwsdbridge.ProjectRepository) RepositoryHealth {
	switch {
	case !repository.LocalPresent:
		return RepositoryMissingLocal
	case !repository.GitPresent:
		return RepositoryMissingGit
	case !repository.RemoteMatches:
		return RepositoryRemoteMismatch
	default:
		return RepositoryReady
	}
}

func liftDiagnostics(source string, diagnostics []mwsdbridge.Diagnostic) []ProjectionDiagnostic {
	out := make([]ProjectionDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		out = append(out, ProjectionDiagnostic{
			Severity: diagnostic.Severity,
			Code:     source + "-" + diagnostic.Code,
			Message:  diagnostic.Message,
		})
	}
	return out
}
