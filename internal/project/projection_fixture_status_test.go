package project

import "github.com/teamswyg/riido-daemon/internal/mwsdbridge"

func sampleSnapshotStatus() mwsdbridge.Status {
	return mwsdbridge.Status{
		Root:                       "/workspace",
		GraphSchemaVersion:         mwsdbridge.GraphSchemaVersion,
		DomainSchemaVersion:        mwsdbridge.DomainSchemaVersion,
		HarnessSchemaVersion:       mwsdbridge.HarnessSchemaVersion,
		OrchestrationSchemaVersion: mwsdbridge.OrchestrationSchemaVersion,
		DocumentCount:              23,
		RepositoryCount:            3,
		DomainName:                 "macmini-workspace",
		HarnessRunCount:            2,
		HarnessNextDirection:       "top-down",
		HarnessRecentDirections:    []string{"top-down", "bottom-up"},
	}
}

func sampleSnapshotDomain() mwsdbridge.DomainExport {
	return mwsdbridge.DomainExport{
		SchemaVersion: mwsdbridge.DomainSchemaVersion,
		Path:          "/workspace/domains/macmini-workspace.lisp",
		Domain:        "macmini-workspace",
	}
}

func sampleSnapshotHarness() mwsdbridge.HarnessIndex {
	return mwsdbridge.HarnessIndex{
		SchemaVersion:    mwsdbridge.HarnessSchemaVersion,
		Path:             "/workspace/harness/runs.jsonl",
		RunCount:         2,
		TopDownCount:     1,
		BottomUpCount:    1,
		LastDirection:    "bottom-up",
		NextDirection:    "top-down",
		RecentDirections: []string{"top-down", "bottom-up"},
	}
}
