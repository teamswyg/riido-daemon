package project

import "github.com/teamswyg/riido-daemon/internal/mwsdbridge"

func sampleSnapshotGraph() mwsdbridge.GraphExport {
	return mwsdbridge.GraphExport{
		SchemaVersion: mwsdbridge.GraphSchemaVersion,
		Root:          "/workspace",
		Documents: []mwsdbridge.Document{
			{
				Path:   "README.md",
				Links:  []string{"docs/GOAL.md"},
				Status: "",
			},
			{
				Path:   "docs/ROADMAP.md",
				ID:     "mws.roadmap",
				Title:  "로드맵",
				Status: "seed",
				Owner:  "local",
			},
			{
				Path:   "docs/GOAL.md",
				ID:     "mws.goal",
				Title:  "목표",
				Status: "seed",
				Owner:  "local",
			},
		},
		Stats: mwsdbridge.GraphStats{
			DocumentCount: 23,
			NodeCount:     23,
			EdgeCount:     100,
		},
	}
}
