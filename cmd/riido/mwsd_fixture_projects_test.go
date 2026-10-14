package main

import "github.com/teamswyg/riido-daemon/internal/mwsdbridge"

func cliMwsdProjects() mwsdbridge.ProjectRegistry {
	return mwsdbridge.ProjectRegistry{
		SchemaVersion:   mwsdbridge.ProjectsSchemaVersion,
		Root:            cliMwsdRoot(),
		RepositoryCount: 1,
		Repositories: []mwsdbridge.ProjectRepository{{
			Name:          "riido-daemon",
			Owner:         "teamswyg",
			Visibility:    "private",
			SSOTScope:     "docs",
			LocalPath:     cliMwsdRoot(),
			Remote:        "https://github.com/teamswyg/riido-daemon",
			Role:          "daemon",
			LocalPresent:  true,
			GitPresent:    true,
			RemoteMatches: true,
		}},
	}
}
