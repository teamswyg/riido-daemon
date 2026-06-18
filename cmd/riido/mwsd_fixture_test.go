package main

import (
	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
)

func cliMwsdSnapshot() mwsdbridge.Snapshot {
	return mwsdbridge.Snapshot{
		Status:        cliMwsdStatus(),
		Graph:         cliMwsdGraph(),
		Domain:        cliMwsdDomain(),
		Harness:       cliMwsdHarness(),
		Orchestration: cliMwsdOrchestration(),
		Projects:      cliMwsdProjects(),
	}
}

func cliMwsdRoot() string {
	return "/tmp/riido-cli-mwsd"
}
