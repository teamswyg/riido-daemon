package project

import "github.com/teamswyg/riido-daemon/internal/mwsdbridge"

func sampleSnapshotProjects() mwsdbridge.ProjectRegistry {
	return mwsdbridge.ProjectRegistry{
		SchemaVersion:   mwsdbridge.ProjectsSchemaVersion,
		Root:            "/workspace",
		DomainPath:      "/workspace/domains/macmini-workspace.lisp",
		RepositoryCount: 3,
		Repositories: []mwsdbridge.ProjectRepository{
			sampleSnapshotRepository("macmini-workspace", "workspace-control-plane", "control-plane", "https://github.com/kimjooyoon/macmini-workspace"),
			sampleSnapshotRepository("gui_engine", "gui-engine", "gui-runtime", "https://github.com/kimjooyoon/gui_engine"),
			sampleSnapshotRepository("riido-daemon", "project-daemon", "project-ssot", "https://github.com/teamswyg/riido-daemon"),
		},
	}
}

func sampleSnapshotRepository(name, scope, role, remote string) mwsdbridge.ProjectRepository {
	return mwsdbridge.ProjectRepository{
		Name:          name,
		Owner:         "kimjooyoon",
		Visibility:    "private",
		SSOTScope:     scope,
		LocalPath:     "/Users/teddy/github/kimjooyoon/" + name,
		Remote:        remote,
		Role:          role,
		LocalPresent:  true,
		GitPresent:    true,
		RemoteMatches: true,
	}
}
