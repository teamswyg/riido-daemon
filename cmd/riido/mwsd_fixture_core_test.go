package main

import "github.com/teamswyg/riido-daemon/internal/mwsdbridge"

func cliMwsdStatus() mwsdbridge.Status {
	return mwsdbridge.Status{
		Root:                       cliMwsdRoot(),
		SocketPath:                 "/tmp/mwsd.sock",
		GraphSchemaVersion:         mwsdbridge.GraphSchemaVersion,
		DomainSchemaVersion:        mwsdbridge.DomainSchemaVersion,
		HarnessSchemaVersion:       mwsdbridge.HarnessSchemaVersion,
		OrchestrationSchemaVersion: mwsdbridge.OrchestrationSchemaVersion,
		DocumentCount:              1,
		RepositoryCount:            1,
	}
}

func cliMwsdDomain() mwsdbridge.DomainExport {
	return mwsdbridge.DomainExport{
		SchemaVersion: mwsdbridge.DomainSchemaVersion,
		Path:          "docs/domain.mws",
		Domain:        "macmini-workspace",
	}
}

func cliMwsdHarness() mwsdbridge.HarnessIndex {
	return mwsdbridge.HarnessIndex{
		SchemaVersion:    mwsdbridge.HarnessSchemaVersion,
		RunCount:         1,
		TopDownCount:     1,
		BottomUpCount:    0,
		LastDirection:    "top-down",
		NextDirection:    "bottom-up",
		RecentDirections: []string{"top-down"},
	}
}
