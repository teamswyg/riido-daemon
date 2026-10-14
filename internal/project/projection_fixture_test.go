package project

import (
	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
)

func sampleSnapshot() mwsdbridge.Snapshot {
	return mwsdbridge.Snapshot{
		Status:        sampleSnapshotStatus(),
		Graph:         sampleSnapshotGraph(),
		Domain:        sampleSnapshotDomain(),
		Harness:       sampleSnapshotHarness(),
		Orchestration: sampleSnapshotOrchestration(),
		Projects:      sampleSnapshotProjects(),
	}
}
