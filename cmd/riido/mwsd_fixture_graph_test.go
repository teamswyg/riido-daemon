package main

import "github.com/teamswyg/riido-daemon/internal/mwsdbridge"

func cliMwsdGraph() mwsdbridge.GraphExport {
	return mwsdbridge.GraphExport{
		SchemaVersion: mwsdbridge.GraphSchemaVersion,
		Root:          cliMwsdRoot(),
		Documents: []mwsdbridge.Document{{
			Path:   "docs/CLI.md",
			ID:     "mws.cli",
			Title:  "CLI migration",
			Status: "in-progress",
			Owner:  "kim",
		}},
		Stats: mwsdbridge.GraphStats{
			DocumentCount: 1,
			NodeCount:     1,
			EdgeCount:     0,
		},
	}
}
